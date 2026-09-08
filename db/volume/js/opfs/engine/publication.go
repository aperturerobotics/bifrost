package engine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/fs"
	"runtime/trace"
	"strconv"
	"strings"
)

const (
	// maxPublicationBytes bounds all prepared index and catalogue output.
	maxPublicationBytes = 32 << 20
	// maxPublicationFiles bounds intent and retirement records.
	maxPublicationFiles = 8192
)

// publication prepares bounded immutable output before acknowledging a root.
// Its caller holds the publication lock until commit returns.
type publication struct {
	// engine owns the backend and validated immutable reads.
	engine *Engine
	// root is the next root, private until the alternate slot is published.
	root *Root
	// output holds bounded new immutable files.
	output []outputFile
	// bytes charges the complete prepared output.
	bytes int
	// retired names old files no longer reachable from this root.
	retired map[string]struct{}
}

// outputFile pairs an immutable name with its prepared bytes.
type outputFile struct {
	// name is unique across attempted publications.
	name string
	// data is written completely before root publication.
	data []byte
}

// newPublication allocates a distinct identity, including after a failed commit.
func newPublication(e *Engine, base *Root) *publication {
	var nonce [16]byte
	// crypto/rand.Read fills the buffer or terminates the process.
	_, _ = rand.Read(nonce[:])
	root := base.CloneVT()
	root.Format = formatVersion
	root.Generation++
	root.Publication = hex.EncodeToString(nonce[:])
	return &publication{engine: e, root: root, retired: make(map[string]struct{})}
}

// add encodes a persistent object into bounded unpublished output.
func (p *publication) add(kind string, value message) (string, error) {
	data, err := encode(value)
	if err != nil {
		return "", err
	}
	return p.addBytes(kind, data)
}

// addBytes retains a bounded immutable file for this publication.
func (p *publication) addBytes(kind string, data []byte) (string, error) {
	if p.bytes+len(data) > maxPublicationBytes || len(p.output) >= maxPublicationFiles {
		return "", ErrLimit
	}
	name := kind + "-" + p.root.Publication + "-" + strconv.Itoa(len(p.output))
	p.output = append(p.output, outputFile{name: name, data: data})
	p.bytes += len(data)
	return name, nil
}

// retire marks a replaced immutable file for later protected deletion.
func (p *publication) retire(name string) {
	p.retired[name] = struct{}{}
}

// commit writes an intent, all immutable files, then one alternate descriptor.
func (p *publication) commit(ctx context.Context) error {
	// Trace the bounded publication through durable root replacement.
	ctx, task := trace.NewTask(ctx, "hydra/opfs-engine/publish")
	defer task.End()
	trace.Logf(ctx, "hydra/opfs-engine/publish/shape", "files=%d bytes=%d retired=%d", len(p.output), p.bytes, len(p.retired))

	// Recover any previous interrupted output before allocating new files.
	if len(p.retired) > maxPublicationFiles {
		return ErrLimit
	}
	if err := p.engine.cleanIntent(ctx); err != nil {
		return err
	}

	// Persist the complete allocation list before creating any orphanable file.
	if len(p.retired) != 0 {
		p.root.RetireThrough = p.root.Generation
	}
	retireName := "retire-" + strconv.FormatUint(p.root.Generation, 10)
	intent := &Files{Publication: p.root.Publication}
	if len(p.retired) != 0 {
		intent.Names = append(intent.Names, retireName)
	}
	for _, output := range p.output {
		intent.Names = append(intent.Names, output.name)
	}
	if err := p.engine.writeMessage(ctx, "intent", intent); err != nil {
		return err
	}
	for _, output := range p.output {
		if err := p.engine.backend.Write(ctx, output.name, output.data); err != nil {
			return err
		}
	}

	// Record retirement before publication so a crash cannot lose cleanup work.
	retired := new(Files)
	for name := range p.retired {
		retired.Names = append(retired.Names, name)
	}
	if len(retired.Names) != 0 {
		if err := p.engine.writeMessage(ctx, retireName, retired); err != nil {
			return err
		}
	}
	if err := p.publishRoot(ctx); err != nil {
		return err
	}

	// Keep newly durable metadata in the existing bounded immutable-file cache.
	// Packs use range-window keys and remain demand-cached at the payload reader.
	p.engine.mtx.Lock()
	if !p.engine.closed {
		for _, output := range p.output {
			if !strings.HasPrefix(output.name, "pack-") {
				p.engine.cacheBytesLocked(output.name, output.data)
			}
		}
	}
	p.engine.mtx.Unlock()
	p.engine.backend.Notify(p.root.Generation)

	// A leftover committed intent is harmless and resolved before the next write.
	_ = p.engine.backend.Remove(ctx, "intent")
	p.engine.mtx.Lock()
	wake := p.engine.wake
	p.engine.mtx.Unlock()
	select {
	case wake <- struct{}{}:
	default:
	}
	return nil
}

// publishRoot excludes root readers only while replacing the durable descriptor.
// Callers hold reclamation protection and the publication lock before this lock.
func (p *publication) publishRoot(ctx context.Context) error {
	unlock, err := p.engine.backend.Lock(ctx, "root", true)
	if err != nil {
		return err
	}
	defer unlock()
	return p.engine.writeMessage(ctx, "root-"+strconv.FormatUint(p.root.Generation%2, 10), p.root)
}

// writeMessage atomically replaces one checksummed persistent record.
func (e *Engine) writeMessage(ctx context.Context, name string, value message) error {
	data, err := encode(value)
	if err != nil {
		return err
	}
	return e.backend.Write(ctx, name, data)
}

// cleanIntent removes only output that was never selected by a valid root.
// The caller holds the publication lock and shared or exclusive reclamation.
func (e *Engine) cleanIntent(ctx context.Context) error {
	// Read the previous intent while the publication lock excludes replacement.
	data, err := e.backend.Read(ctx, "intent", 0, readAll)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	// OPFS creates an empty entry before a writable stream commits. No output
	// can exist until this intent is durable, so an empty entry is discardable.
	if len(data) == 0 {
		return e.backend.Remove(ctx, "intent")
	}
	intent := new(Files)
	if err := decode(data, intent); err != nil {
		return errors.Join(errors.New("decode publication intent (bytes="+strconv.Itoa(len(data))+")"), err)
	}
	if len(intent.Names) > maxPublicationFiles+1 || intent.Publication == "" {
		return ErrCorrupt
	}

	// Keep output selected by the current durable publication.
	root, err := e.loadRoot(ctx)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if root == nil {
		initialized, err := e.hasIdentity(ctx)
		if err != nil {
			return err
		}
		if initialized {
			return ErrCorrupt
		}
	}
	if root == nil || root.Publication != intent.Publication {
		for _, name := range intent.Names {
			if err := e.backend.Remove(ctx, name); err != nil {
				return err
			}
		}
	}
	return e.backend.Remove(ctx, "intent")
}

// Reclaim deletes at most one publication's retired files, then checkpoints.
// It never waits for readers while holding the publication lock.
func (e *Engine) Reclaim(ctx context.Context) (bool, error) {
	// Exclude readers before joining the publication queue.
	release, err := e.backend.Lock(ctx, "reclaim", true)
	if err != nil {
		return false, err
	}
	defer release()
	unlock, err := e.backend.Lock(ctx, "publish", true)
	if err != nil {
		return false, err
	}
	defer unlock()
	root, err := e.loadRoot(ctx)
	if err != nil {
		return false, err
	}
	if root.ReclaimNext > root.RetireThrough {
		return false, nil
	}

	// Advance both root slots before reclaiming the preceding generation.
	if root.ReclaimNext >= root.Generation {
		p := newPublication(e, root)
		return true, p.commit(ctx)
	}

	// Files retired before the preceding generation are absent from both roots.
	name := "retire-" + strconv.FormatUint(root.ReclaimNext, 10)
	data, err := e.backend.Read(ctx, name, 0, readAll)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	if err == nil {
		files := new(Files)
		if err := decode(data, files); err != nil {
			return false, err
		}
		if len(files.Names) > maxPublicationFiles {
			return false, ErrCorrupt
		}
		for _, file := range files.Names {
			if err := e.backend.Remove(ctx, file); err != nil {
				return false, err
			}
		}
		if err := e.backend.Remove(ctx, name); err != nil {
			return false, err
		}
	}

	// Checkpoint completed reclamation through the ordinary root publication.
	p := newPublication(e, root)
	p.root.ReclaimNext++
	return true, p.commit(ctx)
}
