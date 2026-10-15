package sobject_sync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"testing"

	"github.com/aperturerobotics/fastjson"
	"github.com/aperturerobotics/util/ccontainer"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/peer"
)

// leanSyncMarshal encodes real protocol values for primitive observations.
func leanSyncMarshal(t *testing.T, value interface{ MarshalVT() ([]byte, error) }) []byte {
	t.Helper()
	data, err := value.MarshalVT()
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// leanSyncEncoded projects each value as the hex of its encoding, or "nil".
func leanSyncEncoded[T interface {
	comparable
	MarshalVT() ([]byte, error)
}](t *testing.T, a *fastjson.Arena, values []T) *fastjson.Value {
	t.Helper()
	result := a.NewArray()
	for i, value := range values {
		result.SetArrayItem(i, a.NewString(leanSyncHex(t, value)))
	}
	return result
}

// leanSyncHex returns the hex of the encoding of value, or "nil" when it is nil.
func leanSyncHex[T interface {
	comparable
	MarshalVT() ([]byte, error)
}](t *testing.T, value T) string {
	t.Helper()
	var zero T
	if value == zero {
		return "nil"
	}
	return hex.EncodeToString(leanSyncMarshal(t, value))
}

// leanSyncState projects the config and the encoded checkpoint, key epochs,
// operations and invitations. The host merge that combines them is a
// primitive observation, so the model compares them only as values.
func leanSyncState(t *testing.T, a *fastjson.Arena, state *sobject.SOState) *fastjson.Value {
	// Default a missing config.
	t.Helper()
	config := state.GetConfig()
	if config == nil {
		config = &sobject.SharedObjectConfig{}
	}

	// Project each part of the state.
	v := a.NewObject()
	v.Set("config", leanSyncConfig(a, config))
	v.Set("checkpoint", a.NewString(leanSyncHex(t, state.GetCheckpoint())))
	v.Set("epochs", leanSyncEncoded(t, a, state.GetKeyEpochs()))
	v.Set("ops", leanSyncEncoded(t, a, state.GetOps()))
	v.Set("invites", leanSyncEncoded(t, a, state.GetInvites()))
	return v
}

// leanSyncMerged observes the host merge of candidate into previous under the
// candidate's config, returning null when the merge fails.
func leanSyncMerged(t *testing.T, a *fastjson.Arena, objectID string, previous, candidate *sobject.SOState) *fastjson.Value {
	// Merge the candidate into a copy of previous, as the host does.
	t.Helper()
	next := previous.CloneVT()
	next.Config = candidate.GetConfig().CloneVT()
	if checkpoint := candidate.GetCheckpoint(); checkpoint != nil {
		if err := next.AdoptCheckpoint(objectID, checkpoint); err != nil {
			return a.NewNull()
		}
	}
	if err := next.MergeKeyEpochs(objectID, candidate.GetKeyEpochs()); err != nil {
		return a.NewNull()
	}
	for _, op := range candidate.GetOps() {
		if _, err := next.AddOperation(objectID, op); err != nil {
			return a.NewNull()
		}
	}
	if err := next.Validate(objectID); err != nil {
		return a.NewNull()
	}
	return leanSyncState(t, a, next)
}

// TestLeanSyncResponseConformance connects real pages, pinned snapshots and atomic host imports.
func TestLeanSyncResponseConformance(t *testing.T) {
	oracle := leanSyncOracle(t)
	var cases []leanSyncCase
	for seed := range uint64(4) {
		cases = append(cases, leanSyncResponseCases(t, seed)...)
	}
	checkLeanSync(t, oracle, cases)
}

// FuzzLeanSyncResponse varies complete signed responses and host failure observations.
func FuzzLeanSyncResponse(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(19))
	f.Fuzz(func(t *testing.T, seed uint64) {
		checkLeanSync(t, leanSyncOracle(t), leanSyncResponseCases(t, seed))
	})
}

// leanSyncResponseCases compares response preparation and acceptance through their real owners.
func leanSyncResponseCases(t *testing.T, seed uint64) []leanSyncCase {

	// helper.
	t.Helper()
	const objectID = "lean-sync-response"
	owner, reader := mustKeyPair(t), mustKeyPair(t)
	initial := authenticationState(t, objectID, owner, reader)
	initial.Invites = []*sobject.SOInvite{{InviteId: "local capability", TokenHash: []byte("local token")}}
	sender := newAuthenticationPeer(t, objectID, owner, initial)

	// Iterate the test cases.
	for range maxHistoryPageEntries + 1 + int(seed%3) {
		current, err := sender.soHost.GetHostState(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		change, err := sobject.BuildSOConfigChange(objectID, current.Config, current.Config,
			sobject.SOConfigChangeType_SO_CONFIG_CHANGE_TYPE_ADD_INVITE, owner, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := sender.soHost.ApplyConfigChange(t.Context(), change, nil); err != nil {
			t.Fatal(err)
		}
	}
	target, err := sender.soHost.GetHostState(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for range 1 + seed%3 {
		advanceSnapshotCheckpoint(t, objectID, target, owner)
	}
	writeSyncOp(t, objectID, target, owner, "response operation")
	target.Invites = []*sobject.SOInvite{{InviteId: "remote capability", TokenHash: []byte("secret token")}}

	// Record request.
	request := &SOSyncHistoryRequest{Revision: seed + 1, BaseHash: initial.Config.ConfigChainHash}
	response, err := sender.prepareResponse(t.Context(), target, request)
	if err != nil {
		t.Fatal(err)
	}
	changes := response.changes
	digest, err := syncStateHash(target)
	if err != nil {
		t.Fatal(err)
	}

	// Record head.
	head := &SOSyncHead{Revision: request.Revision, ConfigHash: target.Config.ConfigChainHash,
		ConfigSeqno: target.Config.ConfigChainSeqno, StateHash: digest}
	received := &syncReceive{head: head, base: request.BaseHash, cursor: request.BaseHash}
	var cases []leanSyncCase
	var pages []*SOSyncMessage
	for len(response.changes) != 0 {
		message, err := response.nextMessage()
		if err != nil {
			t.Fatal(err)
		}
		var arena fastjson.Arena
		input := arena.NewObject()
		pages = append(pages, message)
		input.Set("op", arena.NewString("appendSyncPage"))
		input.Set("before", leanSyncReceive(t, &arena, received))
		input.Set("page", leanSyncPage(t, &arena, message))
		if err := received.appendPage(message); err != nil {
			t.Fatal(err)
		}
		expected, result := arena.NewObject(), arena.NewObject()
		expected.Set("ok", arena.NewTrue())
		result.Set("ok", arena.NewTrue())
		result.Set("state", leanSyncReceive(t, &arena, received))
		result.Set("recovery", arena.NewFalse())
		expected.Set("received", result)
		cases = append(cases, leanSyncCase{name: "complete response page " + strconv.Itoa(len(cases)),
			request: input.MarshalTo(nil), expected: expected.MarshalTo(nil)})
	}
	message, err := response.nextMessage()
	if err != nil || message.GetSnapshot() == nil {
		t.Fatalf("complete response snapshot: %v", err)
	}

	// The pinned snapshot is the exact encoding whose digest was advertised.
	decoded := &sobject.SOState{}
	if err := decoded.UnmarshalVT(message.GetSnapshot().GetSoState()); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Invites) != 0 {
		t.Fatal("prepared snapshot disclosed local capabilities")
	}
	cases = append(cases, leanSyncReceptionCases(t, head, request.BaseHash, pages)...)
	for variant := range 26 {
		previous, candidate := initial.CloneVT(), decoded.CloneVT()
		receiving := &syncReceive{head: head.CloneVT(), base: append([]byte(nil), received.base...),
			cursor: append([]byte(nil), received.cursor...), size: received.size}
		for _, entry := range changes {
			receiving.changes = append(receiving.changes, entry.CloneVT())
		}
		snapshot := message.GetSnapshot().CloneVT()
		controls := leanSyncHostControls()
		switch variant {
		case 1:
			snapshot.Revision ^= 1
		case 2:
			snapshot.BaseHash = []byte("wrong base")
		case 3:
			receiving.cursor = []byte("incomplete suffix")
		case 6:
			candidate.Config.ConfigChainSeqno++
		case 7:
			candidate.Config.ConfigChainHash = []byte("unadvertised config")
		case 8:
			receiving.changes = receiving.changes[1:]
		case 9:
			receiving.changes[0].Signatures[0].SigData = []byte("invalid signature")
		case 10:
			candidate.Checkpoint.Signatures[0].SigData = []byte("invalid checkpoint proof")
		case 11:
			controls.accessOK = false
		case 12:
			controls.lockOK = false
		case 13:
			controls.writeOK = false
		case 14:
			previous = target.CloneVT()
			previous.Config.ConfigChainSeqno++
		case 15:
			candidate.Config.ConfigChainHash = []byte("equal-sequence fork")
			receiving.head.ConfigHash = candidate.Config.ConfigChainHash
			receiving.cursor = candidate.Config.ConfigChainHash
		case 16:
			previous = decoded.CloneVT()
			receiving.changes = nil
		case 18:
			candidate.Config = nil
		case 19:
			config := candidate.Config.CloneVT()
			config.Participants = config.Participants[:1]
			change, err := sobject.BuildSOConfigChange(objectID, candidate.Config, config,
				sobject.SOConfigChangeType_SO_CONFIG_CHANGE_TYPE_REMOVE_PARTICIPANT, owner, nil)
			if err != nil {
				t.Fatal(err)
			}
			candidate.Config, err = sobject.VerifyConfigChange(objectID, candidate.Config, change)
			if err != nil {
				t.Fatal(err)
			}
			receiving.changes = append(receiving.changes, change)
			receiving.cursor = candidate.Config.ConfigChainHash
			receiving.head.ConfigHash = candidate.Config.ConfigChainHash
			receiving.head.ConfigSeqno = candidate.Config.ConfigChainSeqno
		case 20:
			controls.writeOK = false
			controls.advanced = candidate.CloneVT()
			controls.advanced.Config.ConfigChainSeqno++
		case 21:
			controls.readOK = false
		case 22:
			candidate.Ops = append(candidate.Ops, &sobject.SOOperation{Inner: []byte("malformed operation")})
		case 23:
			candidate.Invites = target.Invites
		case 24:
			previous = decoded.CloneVT()
			advanceSnapshotCheckpoint(t, objectID, previous, owner)
			receiving.changes = nil
		case 25:
			previous = decoded.CloneVT()
			inner, err := previous.GetCheckpointInner()
			if err != nil {
				t.Fatal(err)
			}
			inner = inner.CloneVT()
			inner.StateData = []byte("conflicting checkpoint")
			previous.Checkpoint, err = sobject.BuildSOCheckpoint(owner, inner)
			if err != nil {
				t.Fatal(err)
			}
			receiving.changes = nil
		}
		snapshot.SoState = leanSyncMarshal(t, candidate)
		switch variant {
		case 5:
			snapshot.SoState = []byte{0xff}
		case 17:
			snapshot = nil
		}
		contentDigest := sha256.Sum256(snapshot.GetSoState())
		receiving.head.StateHash = contentDigest[:]
		if variant == 4 {
			receiving.head.StateHash = []byte("wrong digest")
		}
		name := "acceptSyncResponse seed " + strconv.FormatUint(seed, 10) + " variant " + strconv.Itoa(variant)
		cases = append(cases, leanSyncAcceptCase(t, name, objectID, reader, previous, receiving, snapshot, controls))
	}
	cases = append(cases, leanSyncPreparedCases(t, objectID, sender, target, request, changes)...)
	cases = append(cases, leanSyncObsoleteCases(t, objectID, target)...)
	return append(cases, leanSyncBranchCases(t, seed)...)
}

// leanSyncControls injects host failures and reads into one acceptance case.
type leanSyncControls struct {
	// lockOK, accessOK, writeOK and readOK fail the lock, access check, atomic
	// write and state watch when false.
	lockOK, accessOK, writeOK, readOK bool
	// advanced replaces the observed state when the lock is taken.
	advanced *sobject.SOState
	// entry returns retained config changes by hash, or nil.
	entry func(context.Context, string, []byte) (*sobject.SOConfigChange, error)
}

// leanSyncHostControls returns controls under which every host call succeeds.
func leanSyncHostControls() leanSyncControls {
	return leanSyncControls{lockOK: true, accessOK: true, writeOK: true, readOK: true}
}

// leanSyncAcceptCase runs acceptResponse on a host holding previous and
// records the inputs the model reads and the publication the host made.
func leanSyncAcceptCase(t *testing.T, name, objectID string, key crypto.PrivKey, previous *sobject.SOState,
	receiving *syncReceive, snapshot *SOSyncSnapshot, controls leanSyncControls,
) leanSyncCase {
	// Decode the candidate as the receiver does.
	t.Helper()
	candidate := &sobject.SOState{}
	decodeErr := candidate.UnmarshalVT(snapshot.GetSoState())

	// Serve previous through a watch.
	published, wrote := previous, false
	var merge *sobject.SOConfigChange
	observed := ccontainer.NewCContainerVT(previous)
	watch := func(context.Context, string, func()) (ccontainer.Watchable[*sobject.SOState], func(), error) {
		if !controls.readOK {
			return nil, nil, errors.New("injected watch failure")
		}
		return observed, func() {}, nil
	}

	// Lock previous, recording the published state and any merge the host signed.
	lock := func(context.Context, string) (sobject.SOStateLock, error) {
		if controls.advanced != nil {
			observed.SetValue(controls.advanced)
		}
		if !controls.lockOK {
			return nil, errors.New("injected lock failure")
		}
		return sobject.NewSOStateLock(previous, func(_ context.Context, state *sobject.SOState, changes ...*sobject.SOConfigChange) error {
			// Record a signed merge, then publish unless the write fails.
			if len(changes) != 0 && changes[len(changes)-1].GetMerge() != nil {
				merge = changes[len(changes)-1]
			}
			if !controls.writeOK {
				return errors.New("injected atomic write failure")
			}
			published, wrote = state, true
			observed.SetValue(state)
			return nil
		}, func() {}), nil
	}
	host := sobject.NewSOHost(t.Context(), watch, lock, objectID, &sobject.SOHostSyncFuncs{Entry: controls.entry})
	t.Cleanup(host.ClearContext)
	localID, err := peer.IDFromPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	local := NewSOSync(gateLogger(), nil, objectID, localID, key, host, nil,
		func(context.Context, *sobject.SOState) error {
			if !controls.accessOK {
				return errors.New("injected access failure")
			}
			return nil
		})

	// Record the response the receiver holds.
	var arena fastjson.Arena
	digest := sha256.Sum256(snapshot.GetSoState())
	input := arena.NewObject()
	input.Set("object", arena.NewString(objectID))
	input.Set("previous", leanSyncState(t, &arena, previous))
	input.Set("receiving", leanSyncReceive(t, &arena, receiving))
	input.Set("snapshot", leanSyncSnapshot(&arena, snapshot))
	input.Set("digest", arena.NewString(hex.EncodeToString(digest[:])))

	// Record the decoded candidate and its size.
	projection := arena.NewNull()
	if decodeErr == nil {
		projection = leanSyncState(t, &arena, candidate)
	}
	input.Set("decoded", projection)
	input.Set("candidateBytes", arena.NewNumberInt(candidate.SizeVT()))

	// Record the host view before the import, then accept the response.
	beforeRead := arena.NewNull()
	if controls.readOK {
		beforeRead = leanSyncState(t, &arena, previous)
	}
	input.Set("beforeRead", beforeRead)
	input.Set("localPeer", arena.NewString(localID.String()))
	input.Set("lockOK", leanSyncBool(&arena, controls.lockOK))
	input.Set("accessOK", leanSyncBool(&arena, controls.accessOK))
	input.Set("writeOK", leanSyncBool(&arena, controls.writeOK))
	acceptErr := local.acceptResponse(t.Context(), receiving, snapshot)

	// Record the merged candidate, with any merge the host signed.
	merged := arena.NewNull()
	if decodeErr == nil {
		if merge != nil {
			candidate = candidate.CloneVT()
			if candidate.Config, err = sobject.VerifyConfigChange(objectID, candidate.Config, merge); err != nil {
				t.Fatal(err)
			}
		}
		merged = leanSyncMerged(t, &arena, objectID, previous, candidate)
	}
	input.Set("merged", merged)

	// Record the host reads and the state after the import.
	floor, reads := leanSyncBranchReads(t, &arena, host, previous, receiving, merge)
	input.Set("floor", floor)
	input.Set("reads", reads)
	afterRead := arena.NewNull()
	if controls.readOK {
		afterRead = leanSyncState(t, &arena, observed.GetValue())
	}
	input.Set("afterRead", afterRead)

	// Record the result and the publication.
	req, expected, accepted, publication := arena.NewObject(), arena.NewObject(), arena.NewObject(), arena.NewObject()
	req.Set("op", arena.NewString("acceptSyncResponse"))
	req.Set("input", input)
	expected.Set("ok", leanSyncBool(&arena, acceptErr == nil))
	accepted.Set("ok", leanSyncBool(&arena, acceptErr == nil))

	// Record the publication.
	publication.Set("state", leanSyncState(t, &arena, published))
	publication.Set("wrote", leanSyncBool(&arena, wrote))
	publication.Set("revoked", leanSyncBool(&arena, errors.Is(acceptErr, sobject.ErrParticipantRevoked)))
	accepted.Set("host", publication)
	expected.Set("accepted", accepted)
	return leanSyncCase{name: name, request: req.MarshalTo(nil), expected: expected.MarshalTo(nil)}
}

// leanSyncBranchReads projects the host reads the model takes as inputs: the
// checkpoint floor and, for a diverged branch, the held branch above it, the
// record at its cursor and merge, the record the host signed or nil.
func leanSyncBranchReads(t *testing.T, a *fastjson.Arena, host *sobject.SOHost, previous *sobject.SOState,
	receiving *syncReceive, merge *sobject.SOConfigChange,
) (floor, reads *fastjson.Value) {
	// Read the floor the host checks a fork against.
	t.Helper()
	floorSeqno := host.CheckpointFloor(t.Context(), previous)
	floor = a.NewNumberString(strconv.FormatUint(floorSeqno, 10))

	// Read the held branch and the record at its cursor for a diverged branch.
	held, cursorEntry := a.NewNull(), a.NewNull()
	if receiving.branch != nil {
		cursor, branch, err := host.ReadConfigBranch(t.Context(), previous.GetConfig().GetConfigChainHash(), floorSeqno)
		if err == nil {
			// Project the branch with its cursor.
			entries := a.NewArray()
			for i, change := range branch {
				entries.SetArrayItem(i, leanSyncEntry(a, change))
			}
			held = a.NewArray()
			held.SetArrayItem(0, a.NewString(hex.EncodeToString(cursor)))
			held.SetArrayItem(1, entries)

			// Read the record at the cursor, the fork when no held record is shared.
			if entry, err := host.ReadConfigEntry(t.Context(), cursor); err == nil {
				cursorEntry = leanSyncEntry(a, entry)
			}
		}
	}

	// Project the signed merge and the reads.
	signed := a.NewNull()
	if merge != nil {
		signed = leanSyncEntry(a, merge)
	}
	reads = a.NewObject()
	reads.Set("held", held)
	reads.Set("cursorEntry", cursorEntry)
	reads.Set("merge", signed)
	return floor, reads
}

// leanSyncBranchCases accepts responses whose config chain diverged from the
// held one: two owners each extend the genesis apart, and the local owner
// receives the other's branch.
func leanSyncBranchCases(t *testing.T, seed uint64) []leanSyncCase {
	// Two owners hold one genesis and its checkpoint.
	t.Helper()
	objectID := "lean-branch-" + strconv.FormatUint(seed, 10)
	local, remote := mustKeyPair(t), mustKeyPair(t)
	initial := &sobject.SOState{Config: &sobject.SharedObjectConfig{Participants: []*sobject.SOParticipantConfig{
		participantCfg(mustPeerIDStr(t, local), sobject.SOParticipantRole_SOParticipantRole_OWNER),
		participantCfg(mustPeerIDStr(t, remote), sobject.SOParticipantRole_SOParticipantRole_OWNER),
	}}}

	// Retain every record by hash for the host's entry reads.
	history := make(map[string]*sobject.SOConfigChange)
	retain := func(change *sobject.SOConfigChange) {
		hash, err := sobject.HashSOConfigChange(change)
		if err != nil {
			t.Fatal(err)
		}
		history[string(hash)] = change
	}
	entry := func(_ context.Context, _ string, hash []byte) (*sobject.SOConfigChange, error) {
		return history[string(hash)], nil
	}

	// Sign the genesis and its checkpoint.
	genesis, err := sobject.BuildSOConfigChange(objectID, initial.Config, initial.Config, sobject.SOConfigChangeType_SO_CONFIG_CHANGE_TYPE_GENESIS, local, nil)
	if err != nil {
		t.Fatal(err)
	}
	retain(genesis)
	if initial.Config, err = sobject.VerifyConfigChange(objectID, initial.Config, genesis); err != nil {
		t.Fatal(err)
	}
	if initial.Checkpoint, err = sobject.BuildGenesisSOCheckpoint(local, objectID, initial.Config.GetConfigChainHash(), nil); err != nil {
		t.Fatal(err)
	}

	// Apart, the local owner adds one reader and the remote owner adds two.
	extend := func(state *sobject.SOState, signer crypto.PrivKey, count int) (*sobject.SOState, []*sobject.SOConfigChange) {
		// Extend a copy of state by count records.
		t.Helper()
		state = state.CloneVT()
		var changes []*sobject.SOConfigChange
		for range count {
			// Sign a record adding a new reader and apply it.
			next := state.Config.CloneVT()
			next.Participants = append(next.Participants, participantCfg(mustPeerIDStr(t, mustKeyPair(t)), sobject.SOParticipantRole_SOParticipantRole_READER))
			change, err := sobject.BuildSOConfigChange(objectID, state.Config, next, sobject.SOConfigChangeType_SO_CONFIG_CHANGE_TYPE_ADD_PARTICIPANT, signer, nil)
			if err != nil {
				t.Fatal(err)
			}
			retain(change)
			if state.Config, err = sobject.VerifyConfigChange(objectID, state.Config, change); err != nil {
				t.Fatal(err)
			}
			changes = append(changes, change)
		}
		return state, changes
	}
	ours, oursChanges := extend(initial, local, 1)
	theirs, theirsChanges := extend(initial, remote, 2)

	// Each variant receives a branch from the genesis under one fault.
	var cases []leanSyncCase
	for variant := range 8 {
		previous, candidate, changes := ours.CloneVT(), theirs.CloneVT(), theirsChanges
		branch := initial.Config.GetConfigChainHash()
		controls := leanSyncHostControls()
		controls.entry = entry
		switch variant {
		case 1:
			previous, candidate, changes = theirs.CloneVT(), ours.CloneVT(), oursChanges
		case 2:
			previous = initial.CloneVT()
		case 3:
			previous, candidate, changes = theirs.CloneVT(), theirs.CloneVT(), theirsChanges[:1]
			if candidate.Config, err = sobject.VerifyConfigChange(objectID, initial.Config, changes[0]); err != nil {
				t.Fatal(err)
			}
		case 4:
			changes = []*sobject.SOConfigChange{theirsChanges[0].CloneVT(), theirsChanges[1]}
			changes[0].Signatures[0].SigData = []byte("invalid signature")
		case 5:
			branch = []byte("unshared cursor")
		case 6:
			controls.writeOK = false
		case 7:
			controls.accessOK = false
		}

		// Receive the branch as the complete pages would leave it.
		snapshot := &SOSyncSnapshot{SoState: leanSyncMarshal(t, candidate), Revision: 1, BaseHash: previous.Config.GetConfigChainHash()}
		digest := sha256.Sum256(snapshot.GetSoState())
		receiving := &syncReceive{
			head: &SOSyncHead{Revision: 1, ConfigHash: candidate.Config.GetConfigChainHash(),
				ConfigSeqno: candidate.Config.GetConfigChainSeqno(), StateHash: digest[:]},
			base:    snapshot.GetBaseHash(),
			cursor:  candidate.Config.GetConfigChainHash(),
			branch:  branch,
			changes: changes,
		}
		for _, change := range changes {
			receiving.size += change.SizeVT()
		}
		name := "acceptSyncResponse branch seed " + strconv.FormatUint(seed, 10) + " variant " + strconv.Itoa(variant)
		cases = append(cases, leanSyncAcceptCase(t, name, objectID, local, previous, receiving, snapshot, controls))
	}
	return cases
}

// leanSyncPreparedCases exercises the provider read, disclosure projection and complete-frame budget.
func leanSyncPreparedCases(t *testing.T, objectID string, sender *SOSync, target *sobject.SOState,
	request *SOSyncHistoryRequest, changes []*sobject.SOConfigChange,
) []leanSyncCase {
	t.Helper()
	var cases []leanSyncCase
	for variant := range 6 {
		state, req := target.CloneVT(), request.CloneVT()
		var history []*sobject.SOConfigChange
		for _, entry := range changes {
			history = append(history, entry.CloneVT())
		}
		readOK := true
		switch variant {
		case 1:
			readOK = false
		case 2:
			readOK = false
			req.BaseHash = state.Config.ConfigChainHash
		case 3:
			history[0].Signatures[0].SigData = make([]byte, maxMessageSize)
		case 4:
			state.Config.ConfigChainHash = nil
		case 5:
			for range 3 {
				padding := len(history[0].Signatures[0].SigData) + maxMessageSize - 256 - history[0].SizeVT()
				history[0].Signatures[0].SigData = make([]byte, padding)
			}
			if history[0].SizeVT()+256 != maxMessageSize {
				t.Fatal("failed to construct prepareResponse entry boundary")
			}
		}
		host := sobject.NewSOHost(t.Context(), nil, nil, objectID, &sobject.SOHostSyncFuncs{
			History: func(context.Context, string, []byte, []byte) ([]*sobject.SOConfigChange, error) {
				if !readOK {
					return nil, sobject.ErrConfigHistoryUnavailable
				}
				return history, nil
			},
		})
		t.Cleanup(host.ClearContext)
		local := NewSOSync(gateLogger(), nil, objectID, sender.localObjectPeerID, sender.localObjectKey, host, nil)
		wire := state.CloneVT()
		wire.Invites = nil
		encoded := leanSyncMarshal(t, wire)
		snapshot := &SOSyncSnapshot{SoState: encoded, Revision: req.Revision, BaseHash: req.BaseHash}
		frame := &SOSyncMessage{Body: &SOSyncMessage_Snapshot{Snapshot: snapshot}}
		var arena fastjson.Arena
		input, requestJSON := arena.NewObject(), arena.NewObject()
		input.Set("op", arena.NewString("prepareSyncResponse"))
		input.Set("state", leanSyncState(t, &arena, state))
		requestJSON.Set("revision", arena.NewNumberString(strconv.FormatUint(req.Revision, 10)))
		requestJSON.Set("base", arena.NewString(hex.EncodeToString(req.BaseHash)))
		input.Set("request", requestJSON)
		retained := arena.NewNull()
		if readOK {
			retained = leanSyncChanges(t, &arena, history)
		}
		input.Set("history", retained)
		input.Set("encoded", arena.NewString(hex.EncodeToString(encoded)))
		input.Set("bytes", arena.NewNumberInt(frame.SizeVT()))
		response, err := local.prepareResponse(t.Context(), state, req)
		expected := arena.NewObject()
		expected.Set("ok", leanSyncBool(&arena, err == nil))
		projected := arena.NewNull()
		if response != nil {
			projected = leanSyncResponse(t, &arena, response)
		}
		expected.Set("response", projected)
		cases = append(cases, leanSyncCase{name: "prepareSyncResponse variant " + strconv.Itoa(variant),
			request: input.MarshalTo(nil), expected: expected.MarshalTo(nil)})

		input = arena.NewObject()
		input.Set("op", arena.NewString("syncStateHash"))
		input.Set("state", leanSyncState(t, &arena, state))
		input.Set("bytes", arena.NewNumberInt(wire.SizeVT()))
		input.Set("encoded", arena.NewString(hex.EncodeToString(encoded)))
		digest := sha256.Sum256(encoded)
		input.Set("digest", arena.NewString(hex.EncodeToString(digest[:])))
		actual, err := syncStateHash(state)
		expected = arena.NewObject()
		expected.Set("ok", leanSyncBool(&arena, err == nil))
		result := arena.NewNull()
		if err == nil {
			result = arena.NewString(hex.EncodeToString(actual))
		}
		expected.Set("digest", result)
		cases = append(cases, leanSyncCase{name: "syncStateHash variant " + strconv.Itoa(variant),
			request: input.MarshalTo(nil), expected: expected.MarshalTo(nil)})
	}
	return cases
}

// TestLeanSyncResponseSizeLimits measures real snapshot-envelope and state-payload caps.
func TestLeanSyncResponseSizeLimits(t *testing.T) {
	oracle := leanSyncOracle(t)
	const objectID = "lean-response-size"
	owner, reader := mustKeyPair(t), mustKeyPair(t)
	for extra := range 3 {
		state := authenticationState(t, objectID, owner, reader)
		state.Checkpoint.Inner = make([]byte, maxMessageSize)
		snapshot := &SOSyncSnapshot{Revision: 1, BaseHash: state.Config.ConfigChainHash}
		frame := &SOSyncMessage{Body: &SOSyncMessage_Snapshot{Snapshot: snapshot}}
		for range 3 {
			snapshot.SoState = leanSyncMarshal(t, state)
			padding := len(state.Checkpoint.Inner) + maxMessageSize + extra - frame.SizeVT()
			if extra == 2 {
				padding = len(state.Checkpoint.Inner) + maxMessageSize + 1 - state.SizeVT()
			}
			state.Checkpoint.Inner = make([]byte, padding)
		}
		snapshot.SoState = leanSyncMarshal(t, state)
		if extra < 2 && frame.SizeVT() != maxMessageSize+extra || extra == 2 && state.SizeVT() != maxMessageSize+1 {
			t.Fatal("failed to construct snapshot frame boundary")
		}
		local := newAuthenticationPeer(t, objectID, owner, state)
		request := &SOSyncHistoryRequest{Revision: 1, BaseHash: state.Config.ConfigChainHash}
		var arena fastjson.Arena
		input, req := arena.NewObject(), arena.NewObject()
		input.Set("op", arena.NewString("prepareSyncResponse"))
		input.Set("state", leanSyncState(t, &arena, state))
		req.Set("revision", arena.NewNumberInt(1))
		req.Set("base", arena.NewString(hex.EncodeToString(request.BaseHash)))
		input.Set("request", req)
		input.Set("history", arena.NewNull())
		input.Set("encoded", arena.NewString(hex.EncodeToString(snapshot.SoState)))
		input.Set("bytes", arena.NewNumberInt(frame.SizeVT()))
		actual, err := local.prepareResponse(t.Context(), state, request)
		expected := arena.NewObject()
		expected.Set("ok", leanSyncBool(&arena, err == nil))
		result := arena.NewNull()
		if actual != nil {
			result = leanSyncResponse(t, &arena, actual)
		}
		expected.Set("response", result)
		checkLeanSync(t, oracle, []leanSyncCase{{name: "snapshot frame cap plus " + strconv.Itoa(extra),
			request: input.MarshalTo(nil), expected: expected.MarshalTo(nil)}})

		input, expected = arena.NewObject(), arena.NewObject()
		input.Set("op", arena.NewString("syncStateHash"))
		input.Set("state", leanSyncState(t, &arena, state))
		input.Set("bytes", arena.NewNumberInt(state.SizeVT()))
		input.Set("encoded", arena.NewString(hex.EncodeToString(snapshot.SoState)))
		digest := sha256.Sum256(snapshot.SoState)
		input.Set("digest", arena.NewString(hex.EncodeToString(digest[:])))
		actualDigest, err := syncStateHash(state)
		expected.Set("ok", leanSyncBool(&arena, err == nil))
		result = arena.NewNull()
		if err == nil {
			result = arena.NewString(hex.EncodeToString(actualDigest))
		}
		expected.Set("digest", result)
		checkLeanSync(t, oracle, []leanSyncCase{{name: "state payload cap case " + strconv.Itoa(extra),
			request: input.MarshalTo(nil), expected: expected.MarshalTo(nil)}})
	}
}

// leanSyncObsoleteCases covers every ordering of the held and advertised config sequence.
func leanSyncObsoleteCases(t *testing.T, objectID string, state *sobject.SOState) []leanSyncCase {
	t.Helper()
	var cases []leanSyncCase
	for configDelta := -1; configDelta <= 1; configDelta++ {
		head := &SOSyncHead{ConfigSeqno: state.Config.ConfigChainSeqno}
		current := state.CloneVT()
		current.Config.ConfigChainSeqno = uint64(int64(head.ConfigSeqno) + int64(configDelta))
		host, _ := newMemHost(objectID, current)
		t.Cleanup(host.ClearContext)
		local := &SOSync{soHost: host}
		var arena fastjson.Arena
		input, expected := arena.NewObject(), arena.NewObject()
		input.Set("op", arena.NewString("syncResponseObsolete"))
		input.Set("current", leanSyncState(t, &arena, current))
		input.Set("head", leanSyncHead(&arena, head))
		expected.Set("ok", leanSyncBool(&arena, local.responseObsolete(t.Context(), head)))
		cases = append(cases, leanSyncCase{name: "obsolete config delta " + strconv.Itoa(configDelta),
			request: input.MarshalTo(nil), expected: expected.MarshalTo(nil)})
	}
	return cases
}

// leanSyncReceptionCases compares complete and interrupted sequences of the actual received pages.
func leanSyncReceptionCases(t *testing.T, head *SOSyncHead, base []byte, pages []*SOSyncMessage) []leanSyncCase {
	t.Helper()
	var cases []leanSyncCase
	for variant := range 8 {
		receiving := &syncReceive{head: head, base: base, cursor: base}
		var messages []*SOSyncMessage
		for _, page := range pages {
			messages = append(messages, page.CloneVT())
		}
		switch variant {
		case 1:
			messages[0].GetHistoryPage().Revision ^= 1
		case 2:
			messages[1].GetHistoryPage().Cursor = []byte("wrong second cursor")
		case 3:
			messages[0], messages[1] = messages[1], messages[0]
		case 4:
			messages = messages[:len(messages)-1]
		case 5:
			messages = append(messages[:1], messages...)
		case 6:
			messages = append(messages, &SOSyncMessage{Body: &SOSyncMessage_HistoryPage{HistoryPage: &SOSyncHistoryPage{
				Revision: head.Revision, Cursor: head.ConfigHash,
			}}})
		case 7:
			receiving.size = sobject.MaxConfigSuffixBytes
		}
		var arena fastjson.Arena
		input, projected := arena.NewObject(), arena.NewArray()
		input.Set("op", arena.NewString("receiveSyncPages"))
		input.Set("before", leanSyncReceive(t, &arena, receiving))
		for index, message := range messages {
			projected.SetArrayItem(index, leanSyncPage(t, &arena, message))
		}
		input.Set("pages", projected)
		var err error
		for _, message := range messages {
			if err = receiving.appendPage(message); err != nil {
				break
			}
		}
		expected, result := arena.NewObject(), arena.NewNull()
		if err == nil {
			result = leanSyncReceive(t, &arena, receiving)
		}
		expected.Set("ok", leanSyncBool(&arena, err == nil))
		expected.Set("received", result)
		cases = append(cases, leanSyncCase{name: "receiveSyncPages variant " + strconv.Itoa(variant),
			request: input.MarshalTo(nil), expected: expected.MarshalTo(nil)})
	}
	return cases
}
