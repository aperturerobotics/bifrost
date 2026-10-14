//go:build !js

package sobject_world_engine_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/world"
)

// TestCommitForksFromAdmittedWorld holds the publication of one write and
// checks that the next write reads its World and is admitted after it. Each
// Commit returns once its own publication is durable.
func TestCommitForksFromAdmittedWorld(t *testing.T) {
	// Hold the publication of the first write.
	w, so := startGatedWorld(t, t.TempDir(), nil)
	before := admittedOps(t, so)
	applying, release := so.hold(nil)
	first := commitAsync(t, w.newWrite(t, "admit/1"))
	waitReached(t, applying)

	// The second write forks from the first's admitted World.
	tx := w.newWrite(t, "admit/2")
	if found := worldHas(t, tx, "admit/1"); !found {
		t.Fatal("second write does not see the admitted first write")
	}
	second := commitAsync(t, tx)
	waitAdmittedOps(t, so, before+2)

	// Neither Commit returns before the held publication.
	select {
	case err := <-first:
		t.Fatalf("first Commit returned before its receipt: %v", err)
	case err := <-second:
		t.Fatalf("second Commit returned before its receipt: %v", err)
	default:
	}

	// Both return nil once it applies, and the World holds both objects.
	release()
	if err := <-first; err != nil {
		t.Fatalf("first Commit: %v", err)
	}
	if err := <-second; err != nil {
		t.Fatalf("second Commit: %v", err)
	}
	read := newRead(t, w)
	defer read.Discard()
	for _, key := range []string{"admit/1", "admit/2"} {
		if !worldHas(t, read, key) {
			t.Fatalf("%s missing after both Commits", key)
		}
	}
}

// TestCommitReceiptFailure fails the held publication of one write after a
// second write was admitted after it. Both Commits fail, the engine
// reinstalls the World from before them, and the next write builds on that
// World.
func TestCommitReceiptFailure(t *testing.T) {
	// Admit two writes after a durable one, holding the first publication.
	w, so := startGatedWorld(t, t.TempDir(), nil)
	w.create(t, "fail/0")
	before := admittedOps(t, so)
	injected := errors.New("injected publication failure")
	applying, release := so.hold(injected)
	first := commitAsync(t, w.newWrite(t, "fail/1"))
	waitReached(t, applying)
	second := commitAsync(t, w.newWrite(t, "fail/2"))
	waitAdmittedOps(t, so, before+2)

	// Fail it: both Commits fail.
	release()
	if err := <-first; !errors.Is(err, injected) {
		t.Fatalf("first Commit: %v, want the injected failure", err)
	}
	if err := <-second; !errors.Is(err, injected) && !errors.Is(err, block.ErrPublicationDependency) {
		t.Fatalf("second Commit: %v, want the injected failure or its dependency", err)
	}

	// Each failed Commit returns after the engine installed the World from
	// before them, so a rerun forks from it at once.
	read := newRead(t, w)
	for key, found := range map[string]bool{"fail/0": true, "fail/1": false, "fail/2": false} {
		if worldHas(t, read, key) != found {
			t.Fatalf("%s present = %v after the failed Commits, want %v", key, !found, found)
		}
	}
	read.Discard()

	// The next write commits on it.
	w.create(t, "fail/3")
	read = newRead(t, w)
	defer read.Discard()
	for _, key := range []string{"fail/0", "fail/3"} {
		if !worldHas(t, read, key) {
			t.Fatalf("%s missing after the next Commit", key)
		}
	}
}

// TestCommitAdmissionCrash cuts the process after two writes are admitted
// while the first publication is held. A group of publications commits
// atomically, so the restarted engine reads the World from before both, which
// holds the durable write of an earlier process, and commits again.
func TestCommitAdmissionCrash(t *testing.T) {
	const roleEnv = "SPACEWAVE_COMMIT_ADMISSION_CRASH_ROLE"
	const dirEnv = "SPACEWAVE_COMMIT_ADMISSION_CRASH_DIR"
	if role := os.Getenv(roleEnv); role != "" {
		commitAdmissionCrashWorker(t, role, os.Getenv(dirEnv))
		return
	}

	// Run each step in its own process.
	dir := t.TempDir()
	for _, step := range []struct {
		role string
		code int
	}{{"seed", 0}, {"admitted", 74}, {"verify", 0}} {
		ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCommitAdmissionCrash$", "-test.timeout=50s")
		cmd.Env = append(os.Environ(), roleEnv+"="+step.role, dirEnv+"="+dir)
		out, err := cmd.CombinedOutput()
		cancel()
		if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != step.code {
			t.Fatalf("%s: %v\n%s", step.role, err, out)
		}
	}
}

// commitAdmissionCrashWorker runs one step of the admission crash test on the
// s4db storage under dir.
func commitAdmissionCrashWorker(t *testing.T, role, dir string) {
	// Read the SharedObject the seed step created.
	refPath := filepath.Join(dir, "so-ref")
	storage := filepath.Join(dir, "storage")
	var soRef *sobject.SharedObjectRef
	if role != "seed" {
		data, err := os.ReadFile(refPath)
		if err != nil {
			t.Fatal(err)
		}
		soRef = &sobject.SharedObjectRef{}
		if err := soRef.UnmarshalVT(data); err != nil {
			t.Fatal(err)
		}
	}

	switch role {
	// Seed one committed object and save the SharedObject.
	case "seed":
		if err := os.MkdirAll(storage, 0o700); err != nil {
			t.Fatal(err)
		}
		w := startLocalWorld(t, storage, nil)
		w.create(t, "crash/0")
		data, err := w.soRef.MarshalVT()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(refPath, data, 0o600); err != nil {
			t.Fatal(err)
		}

	// Admit two writes behind a held publication and cut.
	case "admitted":
		w, so := startGatedWorld(t, storage, soRef)
		before := admittedOps(t, so)
		applying, _ := so.hold(nil)
		commitAsync(t, w.newWrite(t, "crash/1"))
		waitReached(t, applying)
		commitAsync(t, w.newWrite(t, "crash/2"))
		waitAdmittedOps(t, so, before+2)
		os.Exit(74)

	// Read the World from before both writes, then commit again.
	case "verify":
		w := startLocalWorld(t, storage, soRef)
		read := newRead(t, w)
		want := map[string]bool{"crash/0": true, "crash/1": false, "crash/2": false}
		for key, found := range want {
			if got := worldHas(t, read, key); got != found {
				t.Fatalf("%s: found = %v, want %v", key, got, found)
			}
		}
		read.Discard()
		w.create(t, "crash/next")

	default:
		t.Fatalf("unknown role %q", role)
	}
}

// TestEngineReplacedWithAdmissions stops the engine while two admitted writes
// wait for a held publication, as losing read access does, and serves the World
// again. The new engine starts after the publication fails, reads the World
// from before the writes and commits again.
func TestEngineReplacedWithAdmissions(t *testing.T) {
	// Admit two writes behind a held publication, then stop the engine.
	w, so := startGatedWorld(t, t.TempDir(), nil)
	w.create(t, "replace/0")
	before := admittedOps(t, so)
	injected := errors.New("injected publication failure")
	applying, release := so.hold(injected)
	first := commitAsync(t, w.newWrite(t, "replace/1"))
	waitReached(t, applying)
	second := commitAsync(t, w.newWrite(t, "replace/2"))
	waitAdmittedOps(t, so, before+2)
	so.stop()

	// The replacement waits for the publication, which fails.
	so.serve(t)
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	if _, err := so.ctrl.GetWorldEngine(ctx); err == nil {
		t.Fatal("replacement engine served while the publication was held")
	}
	release()
	eng, err := so.ctrl.GetWorldEngine(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	w.eng = eng
	for _, done := range []<-chan error{first, second} {
		if err := <-done; err == nil {
			t.Fatal("Commit of the stopped engine succeeded")
		}
	}

	// It reads the World from before the writes, and commits on it.
	read := newRead(t, w)
	for key, found := range map[string]bool{"replace/0": true, "replace/1": false, "replace/2": false} {
		if worldHas(t, read, key) != found {
			t.Fatalf("%s present = %v in the replacement engine, want %v", key, !found, found)
		}
	}
	read.Discard()
	w.create(t, "replace/3")
}

// TestCheckpointSkipsPendingAdmission holds the publication of the write that
// makes MinCheckpointOperations operations stable. The engine does not
// checkpoint, or hold the writer lock, while it waits for the receipt, and
// checkpoints once the publication is durable.
func TestCheckpointSkipsPendingAdmission(t *testing.T) {
	// Write one operation short of a checkpoint. Every operation of the only
	// roster member is stable.
	w, so := startGatedWorld(t, t.TempDir(), nil)
	var n int
	for ; admittedOps(t, so) < sobject.MinCheckpointOperations-1; n++ {
		w.create(t, "checkpoint/"+strconv.Itoa(n))
	}

	// Hold the write that reaches it.
	applying, release := so.hold(nil)
	held := commitAsync(t, w.newWrite(t, "checkpoint/"+strconv.Itoa(n)))
	waitReached(t, applying)

	// The writer lock is free and the durable state has no checkpoint.
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	write, err := w.eng.NewTransaction(ctx, true)
	if err != nil {
		t.Fatalf("write while the publication is held: %v", err)
	}
	write.Discard()
	durable, relDurable, err := so.AccessSharedObjectState(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer relDurable()
	checkpoint, err := durable.GetValue().GetCheckpoint(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if height := checkpoint.GetHeight(); height != 0 {
		t.Fatalf("checkpointed at height %d while the publication is held", height)
	}

	// Release it: the engine checkpoints the stable operations.
	release()
	if err := <-held; err != nil {
		t.Fatal(err)
	}
	_, err = durable.WaitValueWithValidator(t.Context(), func(snap sobject.SharedObjectStateSnapshot) (bool, error) {
		return checkpointTrimmed(t.Context(), snap)
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

// commitAsync commits tx in a goroutine and returns the channel its result
// arrives on.
func commitAsync(t *testing.T, tx world.Tx) <-chan error {
	done := make(chan error, 1)
	go func() {
		done <- tx.Commit(t.Context())
	}()
	return done
}

// waitReached waits until a held publication reaches its state head.
func waitReached(t *testing.T, applying <-chan struct{}) {
	t.Helper()
	select {
	case <-applying:
	case <-t.Context().Done():
		t.Fatal(context.Cause(t.Context()))
	}
}

// admittedOps returns the number of operations the admitted state of so holds.
func admittedOps(t *testing.T, so sobject.StatePublisher) int {
	t.Helper()
	ctr, rel, err := so.AccessAdmittedState(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rel()
	set, err := ctr.GetValue().GetOperationSet(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return set.Len()
}

// waitAdmittedOps waits until the admitted state of so holds n operations.
func waitAdmittedOps(t *testing.T, so sobject.StatePublisher, n int) {
	t.Helper()
	ctr, rel, err := so.AccessAdmittedState(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rel()
	_, err = ctr.WaitValueWithValidator(t.Context(), func(snap sobject.SharedObjectStateSnapshot) (bool, error) {
		if snap == nil {
			return false, nil
		}
		set, err := snap.GetOperationSet(t.Context())
		return err == nil && set.Len() >= n, err
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

// newRead opens a read transaction on the installed World of w.
func newRead(t *testing.T, w *localWorld) world.Tx {
	t.Helper()
	tx, err := w.eng.NewTransaction(t.Context(), false)
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

// worldHas reports whether tx holds the object key.
func worldHas(t *testing.T, tx world.Tx, key string) bool {
	t.Helper()
	obj, found, err := tx.GetObject(t.Context(), key)
	world.ReleaseObjectState(obj)
	if err != nil {
		t.Fatal(err)
	}
	return found
}
