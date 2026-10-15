import Spacewave.SObject.State

/-!
# SharedObject host import

Mirrors `SOHost.ImportPeerSnapshot` in `core/sobject/host.go` and
`SOHost.MergePeerBranch` in `core/sobject/host-merge.go`. Ordinary
failure is `none`; committed revocation is a successful `HostResult` with
`revoked = true`, matching Go's write followed by ErrParticipantRevoked.

Byte sizes are the real serialized input sizes. Lock, access callback and
persistence outcomes are explicit inputs. Access validation is assumed
read-only. The provider must publish nothing on a failed write.

`merged` is the observed result of merging the candidate's checkpoint, key
epochs and operations into the held state under the candidate configuration:
`AdoptCheckpoint`, `MergeKeyEpochs`, `AddOperation` and `Validate`, or none
when any of them fails. The model owns the authority decisions around it: the
configuration suffix, revocation, local invitations, access and publication.

The checkpoint floor and the retained-history reads of a branch merge are
observations: `CheckpointFloor`, `ReadConfigBranch` above that floor, and
`ReadConfigEntry` of the held cursor. The signed merge record is the observed
result of `BuildSOConfigMerge` with the local key; the model checks that it
absorbs exactly the held branch after the fork and verifies on the peer's head.
Every configuration is under owner control, so no group control is modeled.
-/

namespace Spacewave.SObject

/-- HostResult separates committed state, revocation notification and publication. -/
structure HostResult where
  state : State
  revoked : Bool
  wrote : Bool
  deriving DecidableEq, Repr

/-- readableBy mirrors the `CanReadState` loop over the participants. -/
def readableBy (config : Config) (peer : String) : Bool :=
  config.participants.any fun p => p.peer == peer &&
    p.role ∈ [Role.reader, Role.writer, Role.owner]

/-- publishHost represents atomic provider publication. -/
def publishHost (next : State) (revoked writeOK : Bool) : Option HostResult :=
  if writeOK then some ⟨next, revoked, true⟩ else none

/-- visibleHost makes unchanged live state explicit for an ordinary rejection. -/
def visibleHost (previous : State) (result : Option HostResult) : HostResult :=
  result.getD ⟨previous, false, false⟩

/--
unappliedEntries drops the proof prefix through the held checkpoint's head when
the proof does not already link to it.
-/
def unappliedEntries (head : String) (entries : List Entry) : List Entry :=
  match entries with
  | [] => []
  | e :: _ =>
    if e.prev = head then entries
    else match entries.findIdx? (·.hash == head) with
      | some i => entries.drop (i + 1)
      | none => entries

/-- A proof that already follows the checkpoint is used unchanged. -/
theorem unappliedEntries_of_follows {cur : Config} {entries : List Entry}
    (h : Follows cur entries) : unappliedEntries cur.hash entries = entries := by
  cases entries with
  | nil => rfl
  | cons e rest => simp [unappliedEntries, h e (by simp)]

/-- Trimming keeps only entries from the original proof. -/
theorem mem_of_mem_unappliedEntries {head : String} {entries : List Entry} {e : Entry}
    (h : e ∈ unappliedEntries head entries) : e ∈ entries := by
  unfold unappliedEntries at h
  split at h
  · simp at h
  · split at h
    · exact h
    · split at h
      · exact List.mem_of_mem_drop h
      · exact h

/-- admitted mirrors `lockPeerImport`: bounds on untrusted work, then the lock. -/
def admitted (candidateBytes entries historyBytes : Nat) (lockOK : Bool) : Bool :=
  candidateBytes ≤ 10 * 1024 * 1024 && entries ≤ 4096 && historyBytes ≤ 8 * 1024 * 1024 && lockOK

/--
commitPeerSnapshot mirrors `commitPeerSnapshot`: revocation, the merge, access
and the write of a candidate whose configuration the caller verified.
-/
def commitPeerSnapshot (previous candidate : State) (entries : List Entry) (localPeer : String)
    (merged : Option State) (accessOK writeOK : Bool) : Option HostResult :=
  if !readableBy candidate.config localPeer then
    if entries.isEmpty then some ⟨previous, true, false⟩
    else
      publishHost {previous with config := candidate.config, epochs := [], ops := []} true writeOK
  else do
    let merged ← merged
    let next := {merged with config := candidate.config, invites := previous.invites}
    if !accessOK then none
    else if next = previous then some ⟨previous, false, false⟩
    else publishHost next false writeOK

/--
belowFloor mirrors `checkMergeFloor`: a suffix that does not follow the held head
starts at a merge fork, which must lie above the checkpoint floor.
-/
def belowFloor (head : String) (floor : Nat) : List Entry → Bool
  | [] => false
  | e :: _ => e.prev != head && e.seqno ≤ floor

/--
importPeerSnapshot mirrors bounds, chain authority for the host's shared object
`object`, the merge floor, revocation, the merge, access and the write.
-/
def importPeerSnapshot (object : String) (previous candidate : State) (entries : List Entry)
    (localPeer : String) (candidateBytes historyBytes floor : Nat) (merged : Option State)
    (lockOK accessOK writeOK : Bool) : Option HostResult :=
  let unapplied := unappliedEntries previous.config.hash entries
  if !admitted candidateBytes entries.length historyBytes lockOK then none
  else if belowFloor previous.config.hash floor unapplied then none
  else if !verifySuffix object previous.config candidate.config unapplied then none
  else commitPeerSnapshot previous candidate unapplied localPeer merged accessOK writeOK

/-! ## Branch merge -/

/--
mergeWins mirrors `MergeWins`: the higher sequence number wins, then the lower
hash. Hashes are equal-length lowercase hex, so string order is byte order.
-/
def mergeWins (seqno : Nat) (hash : String) (otherSeqno : Nat) (otherHash : String) : Bool :=
  if seqno != otherSeqno then otherSeqno < seqno else decide (hash < otherHash)

/--
heldPosition mirrors the hash index of `splitConfigBranches` as the count of held
records through hash: the cursor is 0, and a repeated hash takes its last position.
-/
def heldPosition (cursor : String) (held : List Entry) (hash : String) : Option Nat :=
  held.zipIdx.foldl (fun found (e, i) => if e.hash == hash then some (i + 1) else found)
    (if hash == cursor then some 0 else none)

/--
splitBranches mirrors `splitConfigBranches` before its fork read: the counts of
held and peer records through the newest peer record the held branch holds.
-/
def splitBranches (heldCursor : String) (held : List Entry) (peerCursor : String) (peer : List Entry) :
    Option (Nat × Nat) :=
  peer.zipIdx.foldl (fun found (e, j) => ((heldPosition heldCursor held e.hash).map (·, j + 1)).or found)
    ((heldPosition heldCursor held peerCursor).map (·, 0))

/-- toBranch lists entries as a merge branch. -/
def toBranch (entries : List Entry) : Branch :=
  entries.foldr .cons .nil

/-- BranchReads records the host's observed reads and the signed merge record. -/
structure BranchReads where
  /-- held is `ReadConfigBranch` of the held head above the checkpoint floor. -/
  held : Option (String × List Entry)
  /-- cursorEntry is `ReadConfigEntry` of the held cursor. -/
  cursorEntry : Option Entry
  /-- merge is the record `BuildSOConfigMerge` signs with the local key. -/
  merge : Option Entry
  deriving DecidableEq, Repr

/--
mergeSplit mirrors `MergePeerBranch` after the split at the fork. A peer branch
the held head is part of imports as an ordinary suffix. Otherwise, when the
peer's head wins and the local peer owns it, the local peer commits it with a
merge absorbing the held branch; the peer whose head loses merges instead.
-/
def mergeSplit (object : String) (previous candidate : State) (ours theirs : List Entry)
    (fork merge : Option Entry) (localPeer : String) (merged : Option State) (accessOK writeOK : Bool) :
    Option HostResult :=
  let target := candidate.config
  if theirs.isEmpty then some ⟨previous, false, false⟩
  else if ours.isEmpty then
    if verifySuffix object previous.config target theirs then
      commitPeerSnapshot previous candidate theirs localPeer merged accessOK writeOK
    else none
  else if !mergeWins target.seqno target.hash previous.config.seqno previous.config.hash ||
      !isOwner target localPeer then
    some ⟨previous, false, false⟩
  else
    match fork, merge with
    | some fork, some merge =>
      match fork.config, verifyChange object target merge with
      | some forkConfig, some config =>
        if verifySuffix object (forkConfig.withHead fork.seqno fork.hash) target theirs &&
            merge.merge = .some fork (toBranch ours) then
          commitPeerSnapshot previous {candidate with config} (theirs ++ [merge]) localPeer merged
            accessOK writeOK
        else none
      | _, _ => none
    | _, _ => none

/--
mergePeerBranch mirrors `MergePeerBranch`: bounds, the held branch read, the
split at the fork, and the fork read when the fork is the held cursor.
-/
def mergePeerBranch (object : String) (previous candidate : State) (cursor : String) (entries : List Entry)
    (localPeer : String) (candidateBytes historyBytes : Nat) (reads : BranchReads) (merged : Option State)
    (lockOK accessOK writeOK : Bool) : Option HostResult :=
  if !admitted candidateBytes entries.length historyBytes lockOK then none
  else
    match reads.held with
    | none => none
    | some (heldCursor, held) =>
      match splitBranches heldCursor held cursor entries with
      | none => none
      | some (heldAt, peerAt) =>
        let fork := if 0 < heldAt then held[heldAt - 1]? else reads.cursorEntry.filter (·.hash == heldCursor)
        mergeSplit object previous candidate (held.drop heldAt) (entries.drop peerAt) fork reads.merge
          localPeer merged accessOK writeOK

/-! ## Publication -/

/-- Ordinary rejection never publishes a working copy. -/
theorem visibleHost_rejected (previous : State) :
    visibleHost previous none = ⟨previous, false, false⟩ := rfl

/-- Provider failure cannot publish an accepted working copy under its atomic contract. -/
theorem publishHost_failure (next : State) (revoked : Bool) :
    publishHost next revoked false = none := rfl

/-- Publication retains exactly the prepared state and notification. -/
theorem publishHost_spec {next : State} {revoked writeOK : Bool} {out : HostResult}
    (h : publishHost next revoked writeOK = some out) :
    out.state = next ∧ out.revoked = revoked ∧ out.wrote = true := by
  unfold publishHost at h
  split at h
  · cases h; exact ⟨rfl, rfl, rfl⟩
  · contradiction

/-- Every accepted suffix is also a sequence of individually verified changes. -/
theorem applySuffix_verified {obj : String} {es : List Entry} {cur next : Config}
    (h : applySuffix obj cur es = some next) : es.foldlM (verifyChange obj) cur = some next := by
  induction es generalizing cur with
  | nil => exact h
  | cons e es ih =>
    simp only [applySuffix, List.foldlM_cons] at h ⊢
    cases first : suffixStep obj cur e with
    | none => simp [first] at h
    | some mid =>
      simp only [first, Option.bind_eq_bind, Option.bind_some] at h
      have checked : verifyChange obj cur e = some mid := by
        unfold suffixStep at first
        split at first
        · exact first
        · contradiction
      simp only [checked, Option.bind_eq_bind, Option.bind_some]
      exact ih h

/--
A verified suffix never lowers the sequence number, given nonempty entry hashes:
one that follows the checkpoint advances it by its length, and one from a merge
fork must end above it.
-/
theorem verifySuffix_seqno {obj : String} {previous candidate : Config} {entries : List Entry}
    (hashes : ∀ e ∈ entries, e.hash ≠ "")
    (h : verifySuffix obj previous candidate entries = true) :
    previous.seqno ≤ candidate.seqno := by
  simp only [verifySuffix, Bool.and_eq_true] at h
  obtain ⟨⟨checkpoint, _⟩, suffix⟩ := h
  have checkpoint' : previous.hash ≠ "" := by simpa using checkpoint
  split at suffix
  · rename_i next applied
    simp only [Bool.and_eq_true, Bool.or_eq_true, decide_eq_true_eq, Config.same, beq_iff_eq,
      List.isPerm_iff] at suffix
    obtain ⟨linked | above, ⟨_, same⟩, _⟩ := suffix
    · have follows : Follows previous entries := by
        intro e he
        cases entries with
        | nil => simp at he
        | cons x xs =>
          simp only [List.head?_cons, Option.mem_def, Option.some.injEq] at he
          subst he
          simpa using linked
      rw [suffixBase_follows follows, Option.bind_some] at applied
      have progressed := foldlM_verifyChange_seqno checkpoint' hashes (applySuffix_verified applied)
      omega
    · omega
  · contradiction

/-! ## Peer import -/

/-- A committed write publishes the candidate configuration with the local invitations. -/
theorem commitPeerSnapshot_target {previous candidate : State} {entries : List Entry} {peer : String}
    {merged : Option State} {accessOK writeOK : Bool} {out : HostResult}
    (h : commitPeerSnapshot previous candidate entries peer merged accessOK writeOK = some out) :
    out.state = previous ∨
      (out.state.config = candidate.config ∧ out.state.invites = previous.invites) := by
  unfold commitPeerSnapshot at h
  split at h
  · split at h
    · cases h; exact Or.inl rfl
    · obtain ⟨state, _, _⟩ := publishHost_spec h
      exact Or.inr ⟨by rw [state], by rw [state]⟩
  · cases m : merged with
    | none => simp [m] at h
    | some m' =>
      simp only [m, Option.bind_eq_bind, Option.bind_some] at h
      split at h
      · contradiction
      · split at h
        · cases h; exact Or.inl rfl
        · obtain ⟨state, _, _⟩ := publishHost_spec h
          exact Or.inr ⟨by rw [state], by rw [state]⟩

/-- Every accepted import was authorized by the configuration suffix from the held head. -/
theorem importPeerSnapshot_authority {object : String}
    {previous candidate : State} {entries : List Entry} {peer : String} {bytes history floor : Nat}
    {merged : Option State} {lockOK accessOK writeOK : Bool} {out : HostResult}
    (h : importPeerSnapshot object previous candidate entries peer bytes history floor merged
      lockOK accessOK writeOK = some out) :
    verifySuffix object previous.config candidate.config
      (unappliedEntries previous.config.hash entries) = true := by
  unfold importPeerSnapshot at h
  split at h
  · contradiction
  · dsimp only at h
    split at h
    · contradiction
    · split at h
      · contradiction
      · rename_i authorized
        simpa using authorized

/--
An import keeps the held state, or commits the candidate configuration with the
local invitations.
-/
theorem importPeerSnapshot_target {object : String}
    {previous candidate : State} {entries : List Entry} {peer : String} {bytes history floor : Nat}
    {merged : Option State} {lockOK accessOK writeOK : Bool} {out : HostResult}
    (h : importPeerSnapshot object previous candidate entries peer bytes history floor merged
      lockOK accessOK writeOK = some out) :
    out.state = previous ∨
      (out.state.config = candidate.config ∧ out.state.invites = previous.invites) := by
  unfold importPeerSnapshot at h
  split at h
  · contradiction
  · dsimp only at h
    split at h
    · contradiction
    · split at h
      · contradiction
      · exact commitPeerSnapshot_target h

/-- No accepted peer import can roll back the configuration sequence. -/
theorem importPeerSnapshot_config_monotone {object : String}
    {previous candidate : State} {entries : List Entry} {peer : String} {bytes history floor : Nat}
    {merged : Option State} {lockOK accessOK writeOK : Bool} {out : HostResult}
    (hashes : ∀ e ∈ entries, e.hash ≠ "")
    (h : importPeerSnapshot object previous candidate entries peer bytes history floor merged
      lockOK accessOK writeOK = some out) :
    previous.config.seqno ≤ out.state.config.seqno := by
  have seq := verifySuffix_seqno (fun e mem => hashes e (mem_of_mem_unappliedEntries mem))
    (importPeerSnapshot_authority h)
  rcases importPeerSnapshot_target h with unchanged | ⟨advanced, _⟩
  · simp [unchanged]
  · rw [advanced]; exact seq

/-- A suffix that follows the held head is never below the merge floor. -/
theorem belowFloor_follows {cur : Config} {entries : List Entry} (floor : Nat)
    (h : Follows cur entries) : belowFloor cur.hash floor entries = false := by
  cases entries with
  | nil => rfl
  | cons e rest => simp [belowFloor, h e (by simp)]

/-- A proved removal commits without the merge or the access callback. -/
theorem importPeerSnapshot_revoked {object : String}
    {previous candidate : State} {entries : List Entry} {peer : String} {bytes history : Nat}
    (floor : Nat) (merged : Option State) (accessOK : Bool)
    (bounded : bytes ≤ 10 * 1024 * 1024 ∧ entries.length ≤ 4096 ∧
      history ≤ 8 * 1024 * 1024)
    (follows : Follows previous.config entries)
    (chain : verifySuffix object previous.config candidate.config entries = true)
    (removed : readableBy candidate.config peer = false) (nonempty : entries ≠ []) :
    importPeerSnapshot object previous candidate entries peer bytes history floor merged true accessOK
      true =
      some ⟨{previous with config := candidate.config, epochs := [], ops := []}, true, true⟩ := by
  simp [importPeerSnapshot, admitted, bounded.1, bounded.2.1, bounded.2.2,
    unappliedEntries_of_follows follows, belowFloor_follows floor follows, chain, commitPeerSnapshot,
    removed, nonempty, publishHost]

/-- A readable import publishes the merge with the candidate configuration and local invitations. -/
theorem importPeerSnapshot_merged {object : String}
    {previous candidate merged : State} {entries : List Entry} {peer : String} {bytes history floor : Nat}
    (bounded : bytes ≤ 10 * 1024 * 1024 ∧ entries.length ≤ 4096 ∧
      history ≤ 8 * 1024 * 1024)
    (follows : Follows previous.config entries)
    (chain : verifySuffix object previous.config candidate.config entries = true)
    (readable : readableBy candidate.config peer = true)
    (changed : {merged with config := candidate.config, invites := previous.invites} ≠ previous) :
    importPeerSnapshot object previous candidate entries peer bytes history floor (some merged) true true
      true =
      some ⟨{merged with config := candidate.config, invites := previous.invites}, false, true⟩ := by
  simp [importPeerSnapshot, admitted, bounded.1, bounded.2.1, bounded.2.2,
    unappliedEntries_of_follows follows, belowFloor_follows floor follows, chain, commitPeerSnapshot,
    readable, changed, publishHost]

/--
A branch merge keeps the held state, or publishes a configuration it verified:
the peer's suffix from the held head, or a merge record on the peer's head,
which the local peer owns.
-/
theorem mergeSplit_authority {object : String} {previous candidate : State} {ours theirs : List Entry}
    {fork merge : Option Entry} {peer : String} {merged : Option State} {accessOK writeOK : Bool}
    {out : HostResult}
    (h : mergeSplit object previous candidate ours theirs fork merge peer merged accessOK writeOK = some out) :
    out.state = previous ∨
      (verifySuffix object previous.config candidate.config theirs = true ∧
        out.state.config = candidate.config) ∨
      (∃ entry, verifyChange object candidate.config entry = some out.state.config ∧
        isOwner candidate.config peer = true ∧
        mergeWins candidate.config.seqno candidate.config.hash previous.config.seqno previous.config.hash = true) := by
  unfold mergeSplit at h
  dsimp only at h
  split at h
  · cases h; exact Or.inl rfl
  split at h
  · split at h
    · rename_i verified
      rcases commitPeerSnapshot_target h with same | ⟨config, _⟩
      · exact Or.inl same
      · exact Or.inr (Or.inl ⟨verified, config⟩)
    · contradiction
  split at h
  · cases h; exact Or.inl rfl
  rename_i notLosing
  have owner : isOwner candidate.config peer = true := by simp_all
  have wins : mergeWins candidate.config.seqno candidate.config.hash previous.config.seqno
      previous.config.hash = true := by simp_all
  split at h
  · split at h
    · rename_i verified
      split at h
      · rcases commitPeerSnapshot_target h with same | ⟨mergedConfig, _⟩
        · exact Or.inl same
        · exact Or.inr (Or.inr ⟨_, by rw [mergedConfig]; exact verified, owner, wins⟩)
      · contradiction
    · contradiction
  · contradiction

/-- mergePeerBranch inherits the authority of mergeSplit. -/
theorem mergePeerBranch_authority {object : String} {previous candidate : State} {cursor : String}
    {entries : List Entry} {peer : String} {bytes history : Nat} {reads : BranchReads}
    {merged : Option State} {lockOK accessOK writeOK : Bool} {out : HostResult}
    (h : mergePeerBranch object previous candidate cursor entries peer bytes history reads merged
      lockOK accessOK writeOK = some out) :
    out.state = previous ∨
      (∃ theirs, (∀ e ∈ theirs, e ∈ entries) ∧
        verifySuffix object previous.config candidate.config theirs = true ∧
        out.state.config = candidate.config) ∨
      (∃ entry, verifyChange object candidate.config entry = some out.state.config ∧
        isOwner candidate.config peer = true ∧
        mergeWins candidate.config.seqno candidate.config.hash previous.config.seqno previous.config.hash = true) := by
  unfold mergePeerBranch at h
  split at h
  · contradiction
  split at h
  · contradiction
  split at h
  · contradiction
  rcases mergeSplit_authority h with same | ⟨verified, config⟩ | signed
  · exact Or.inl same
  · exact Or.inr (Or.inl ⟨_, fun _ mem => List.mem_of_mem_drop mem, verified, config⟩)
  · exact Or.inr (Or.inr signed)

/-- The winning head is never below the losing one. -/
theorem mergeWins_le {seqno other : Nat} {hash otherHash : String}
    (h : mergeWins seqno hash other otherHash = true) : other ≤ seqno := by
  unfold mergeWins at h
  split at h <;> simp_all <;> omega

/--
A branch merge cannot roll the configuration back: a peer suffix advances the
held head, and a merge record extends the winning head.
-/
theorem mergePeerBranch_config_monotone {object : String} {previous candidate : State} {cursor : String}
    {entries : List Entry} {peer : String} {bytes history : Nat} {reads : BranchReads}
    {merged : Option State} {lockOK accessOK writeOK : Bool} {out : HostResult}
    (hashes : ∀ e ∈ entries, e.hash ≠ "") (head : candidate.config.hash ≠ "")
    (h : mergePeerBranch object previous candidate cursor entries peer bytes history reads merged
      lockOK accessOK writeOK = some out) :
    previous.config.seqno ≤ out.state.config.seqno := by
  rcases mergePeerBranch_authority h with same | ⟨_, within, verified, config⟩ | ⟨_, verified, _, wins⟩
  · simp [same]
  · rw [config]
    exact verifySuffix_seqno (fun e mem => hashes e (within e mem)) verified
  · obtain ⟨_, _, next, _, _, _, seqno, _, _⟩ := verifyChange_spec verified
    have le := mergeWins_le wins
    simp [next, Config.withHead, seqno, expectedSeqno, head] at *
    omega

end Spacewave.SObject
