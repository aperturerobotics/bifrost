/-!
# SharedObject configuration chain

Models the configuration chain decisions in `core/sobject/config-chain.go` and
`SharedObjectConfig.Validate` in `core/sobject/sobject.go`, checked against
Spacewave `cc4d9007e` plus the owner-retention fix:

- `Config.validate` mirrors `SharedObjectConfig.Validate`.
- `verifySignatures` mirrors `verifyConfigChangeSignatures`.
- `validateSelfEnroll` mirrors `validateSelfEnrollPeerChange`.
- `verifyChange` mirrors `VerifyConfigChange`.
- `mergeValid` mirrors `verifyConfigMerge` and `mergeConfigBranch`.
- `suffixBase` mirrors `ConfigSuffixBase`.
- `verifySuffix` mirrors `VerifyConfigChainSuffix`.
- `verifyChain` mirrors `VerifyConfigChain`.

Abstraction. Peer IDs, entity IDs and hashes are opaque strings. The Go
projection maps a peer ID that fails `ParsePeerID` to the empty string and
writes each hash as lowercase hex, so a 32-byte digest is 64 characters. An
entry carries its own hash, computed by Go over the entry without its
signatures, and the shared object it is bound to. A signature is its signer,
empty when its key does not parse, and whether it verified over the same bytes;
unforgeability is the assumption that `valid` holds only for the named signer's
key. Every verifier takes the shared object it verifies for. Role and change
type codes stay integers because protobuf decoding keeps values outside the
enum. Every configuration is under owner control: the model has no group
control, sequencer or sealed checkpoint, so a merge merges participants only.
-/

namespace Spacewave.SObject

/-- Role is a `SOParticipantRole` code. -/
abbrev Role := Int

namespace Role

def unknown : Role := 0
def reader : Role := 1
def writer : Role := 2
def owner : Role := 3

end Role

/-- ChangeType is a `SOConfigChangeType` code. -/
abbrev ChangeType := Int

namespace ChangeType

def genesis : ChangeType := 1
def addParticipant : ChangeType := 2
def removeParticipant : ChangeType := 3
def addInvite : ChangeType := 4
def revokeInvite : ChangeType := 5
def incrementInviteUses : ChangeType := 6
def selfEnrollPeer : ChangeType := 7
def transferOwnership : ChangeType := 8
def setRoster : ChangeType := 9
def setSequencer : ChangeType := 10
def setControl : ChangeType := 11
def merge : ChangeType := 12

/--
peerChanges mirrors `isPeerConfigChange`: the change types a peer may relay in
a suffix or a merge may absorb.
-/
def peerChanges : List ChangeType :=
  [addParticipant, removeParticipant, addInvite, revokeInvite, incrementInviteUses,
    transferOwnership, setRoster, setSequencer, setControl, merge]

end ChangeType

/-- seqnoLimit bounds a `uint64` configuration sequence number. -/
def seqnoLimit : Nat := 2 ^ 64

/-- digestLength is the hex length of a 32-byte SHA-256 digest. -/
def digestLength : Nat := 64

/-- Participant is one `SOParticipantConfig`. -/
structure Participant where
  peer : String
  role : Role
  entity : String
  deriving DecidableEq, Repr

/-- Config is a `SharedObjectConfig` with its applied chain head. -/
structure Config where
  participants : List Participant
  hash : String
  seqno : Nat
  deriving DecidableEq, Repr

/-- Sig is the verification outcome of a config change signature. -/
structure Sig where
  signer : String
  valid : Bool
  deriving DecidableEq, Repr

mutual

/--
Entry is one `SOConfigChange`; `config` is none when Go's field is nil, and
`merge` is its `SOConfigMerge`.
-/
structure Entry where
  object : String
  seqno : Nat
  config : Option Config
  sigs : List Sig
  prev : String
  kind : ChangeType
  hash : String
  merge : Merge

/-- Merge is an absent `SOConfigMerge`, or its fork and the branch after it. -/
inductive Merge where
  | none
  | some (fork : Entry) (branch : Branch)

/-- Branch is the records a merge absorbs, oldest first. -/
inductive Branch where
  | nil
  | cons (head : Entry) (tail : Branch)

end

deriving instance DecidableEq, Repr for Entry, Merge, Branch

/-- Participant.valid mirrors `SOParticipantConfig.Validate`. -/
def Participant.valid (p : Participant) : Bool :=
  p.peer != "" && Role.reader ≤ p.role && p.role ≤ Role.owner

/-- Config.hasOwner reports whether any participant is an OWNER. -/
def Config.hasOwner (c : Config) : Bool :=
  c.participants.any (·.role == Role.owner)

/--
Config.validate mirrors `SharedObjectConfig.Validate`. An empty participant
list is valid only as a signed terminal head; a nonempty one needs valid,
unique participants and an OWNER.
-/
def Config.validate (c : Config) : Bool :=
  if c.participants.isEmpty then
    c.hash.length == digestLength
  else
    c.participants.all Participant.valid &&
      decide (c.participants.map (·.peer)).Nodup &&
      c.hasOwner

/-- Config.withHead mirrors `configWithAppliedConfigChainHead`. -/
def Config.withHead (c : Config) (seqno : Nat) (hash : String) : Config :=
  { c with seqno, hash }

/--
Config.same mirrors `EqualSOConfigs`: equal metadata and the same participants
in any order. Go sorts by peer ID, which matches a permutation whenever one
side has unique peers, as every validated configuration does.
-/
def Config.same (a b : Config) : Bool :=
  a.hash == b.hash && a.seqno == b.seqno &&
    a.participants.isPerm b.participants

/-- isOwner mirrors `isOwnerPeer`. -/
def isOwner (c : Config) (peer : String) : Bool :=
  c.participants.any fun p => p.peer == peer && p.role == Role.owner

/-- lookupPeer mirrors a Go map built from participants: the last match wins. -/
def lookupPeer (ps : List Participant) (peer : String) : Option Participant :=
  ps.reverse.find? (·.peer == peer)

/-- entityRole mirrors `participantRoleForEntity`. -/
def entityRole (c : Config) (entity : String) : Role :=
  c.participants.foldl (fun r p => if p.entity == entity && p.role > r then p.role else r)
    Role.unknown

/--
validateSelfEnroll mirrors `validateSelfEnrollPeerChange`: metadata is
unchanged, every existing participant is preserved, and exactly one new
participant is added, the signer, bound to an existing entity with no more
than that entity's strongest role.
-/
def validateSelfEnroll (cur next : Config) (signer : String) : Bool :=
  let added := next.participants.filter fun p => !cur.participants.any (·.peer == p.peer)
  next.hash == cur.hash && next.seqno == cur.seqno &&
    next.participants.length == cur.participants.length + 1 &&
    cur.participants.all (fun p => lookupPeer next.participants p.peer ==
      lookupPeer cur.participants p.peer) &&
    cur.participants.all (·.peer != signer) &&
    match added with
    | [a] =>
      a.peer == signer && a.entity != "" && entityRole cur a.entity != Role.unknown &&
        a.role ≤ entityRole cur a.entity
    | _ => false

/--
signerAuthorized reports whether `cfg` lets `signer` sign the entry: as an
OWNER, or as the peer an admissible self-enrollment adds.
-/
def signerAuthorized (e : Entry) (cfg : Config) (signer : String) : Bool :=
  if e.kind == ChangeType.selfEnrollPeer then
    match e.config with
    | some next => validateSelfEnroll cfg next signer
    | none => false
  else
    isOwner cfg signer

/--
verifySignatures mirrors `verifyConfigChangeSignatures`: at least one
signature, exactly one for a self-enrollment, from distinct parsed signers,
each authorized by `cfg` and verified.
-/
def verifySignatures (e : Entry) (cfg : Config) : Bool :=
  !e.sigs.isEmpty && (e.kind != ChangeType.selfEnrollPeer || e.sigs.length == 1) &&
    decide (e.sigs.map (·.signer)).Nodup &&
    e.sigs.all fun s => s.signer != "" && s.valid && signerAuthorized e cfg s.signer

/-- expectedSeqno is the sequence number that follows the head of `cur`. -/
def expectedSeqno (cur : Config) : Nat :=
  if cur.hash = "" then 0 else cur.seqno + 1

/--
mergeParticipants mirrors `mergeByKey` over participants: it applies the
change from `parent` to `next` onto `target` by peer. Go replaces a changed
participant in place, and this appends it, so the two results are permutations
whenever `target` has unique peers.
-/
def mergeParticipants (target parent next : List Participant) : List Participant :=
  let changed := next.filter fun q => lookupPeer parent q.peer != some q
  let removed (peer : String) := parent.any (·.peer == peer) && !next.any (·.peer == peer)
  target.filter (fun p => !removed p.peer && !changed.any (·.peer == p.peer)) ++ changed

/-- Delta is one absorbed record: its signers and the configurations around it. -/
structure Delta where
  signers : List String
  parent : Config
  next : Config
  deriving DecidableEq, Repr

/--
absorb mirrors one `mergeConfigBranch` step: a record every signer of which
owns the merged configuration so far applies its change, when the result stays
valid.
-/
def absorb (result : Config) (d : Delta) : Config :=
  let candidate := { result with
    participants := mergeParticipants result.participants d.parent.participants d.next.participants }
  if !d.signers.isEmpty && d.signers.all (isOwner result) && candidate.validate then candidate
  else result

mutual

/--
verifyChange mirrors `VerifyConfigChange`. It accepts an entry bound to
`obj`, linked to the head of `cur`, at the next sequence number, authorized by
`cur`, whose configuration with the entry as its head is valid and, for a
merge, is the one its branch yields on `cur`, and returns that configuration.
-/
def verifyChange (obj : String) (cur : Config) : Entry → Option Config
  | e@⟨_, _, config, _, _, _, _, merge⟩ =>
    match config with
    | none => none
    | some next =>
      let result := next.withHead e.seqno e.hash
      if e.object = obj ∧ e.prev = cur.hash ∧ expectedSeqno cur < seqnoLimit ∧
          e.seqno = expectedSeqno cur ∧ verifySignatures e cur ∧ result.validate then
        if mergeValid obj cur e.kind next merge then some result else none
      else
        none

/--
mergeValid mirrors `verifyConfigMerge`: only a merge carries a branch, and a
merge must carry the configuration `next` that its branch yields on `cur`.
-/
def mergeValid (obj : String) (cur : Config) (kind : ChangeType) (next : Config) :
    Merge → Bool
  | .none => kind != ChangeType.merge
  | .some fork branch =>
    kind == ChangeType.merge &&
      match fork.config with
      | none => false
      | some forkConfig =>
        match branchDeltas obj (forkConfig.withHead fork.seqno fork.hash) branch with
        | some (d :: ds) => next.same ((d :: ds).foldl absorb cur)
        | _ => false

/--
branchDeltas mirrors the verification loop of `mergeConfigBranch`: each
absorbed record is a peer change verified on the configuration before it.
-/
def branchDeltas (obj : String) (parent : Config) : Branch → Option (List Delta)
  | .nil => some []
  | .cons e rest =>
    if e.kind ∈ ChangeType.peerChanges then
      match verifyChange obj parent e with
      | none => none
      | some next => (branchDeltas obj next rest).map (⟨e.sigs.map (·.signer), parent, next⟩ :: ·)
    else
      none

end

/-- suffixStep is one `VerifyConfigChainSuffix` step: a peer change type, then `verifyChange`. -/
def suffixStep (obj : String) (cur : Config) (e : Entry) : Option Config :=
  if e.kind ∈ ChangeType.peerChanges then verifyChange obj cur e else none

/-- applySuffix folds `suffixStep` over entries from a checkpoint. -/
def applySuffix (obj : String) (cur : Config) (es : List Entry) : Option Config :=
  es.foldlM (suffixStep obj) cur

mutual

/--
absorbedFork mirrors `AbsorbedFork`: the fork of the merge in `e` that absorbs
the record hashed `head`, searching merges nested in absorbed branches.
-/
def absorbedFork (head : String) : Entry → Option Entry
  | ⟨_, _, _, _, _, kind, _, .some fork branch⟩ =>
    if kind == ChangeType.merge && head != "" then branchFork head fork branch else none
  | ⟨_, _, _, _, _, _, _, .none⟩ => none

/-- branchFork searches one merge's branch for `head`, returning `fork` for a direct record. -/
def branchFork (head : String) (fork : Entry) : Branch → Option Entry
  | .nil => none
  | .cons e rest => if e.hash == head then some fork else (absorbedFork head e).or (branchFork head fork rest)

end

/--
suffixBase mirrors `ConfigSuffixBase`: a suffix that does not follow `cur`
starts at the fork of the first merge that absorbs the head of `cur`.
-/
def suffixBase (cur : Config) : List Entry → Option Config
  | [] => some cur
  | es@(first :: _) =>
    if first.prev = cur.hash then some cur
    else
      match es.findSome? (absorbedFork cur.hash) with
      | some fork =>
        if fork.hash = first.prev then fork.config.map (·.withHead fork.seqno fork.hash) else none
      | none => none

/--
verifySuffix mirrors `VerifyConfigChainSuffix`: from a nonempty valid
checkpoint, the entries, starting at its head or at the fork of a merge that
absorbs it, must reach a configuration equal to the candidate. A suffix from a
fork must end above the checkpoint.
-/
def verifySuffix (obj : String) (cur cand : Config) (es : List Entry) : Bool :=
  cur.hash != "" && cur.validate &&
    match (suffixBase cur es).bind (applySuffix obj · es) with
    | some c => (es.head?.all (·.prev == cur.hash) || cur.seqno < c.seqno) && c.same cand
    | none => false

/-- genesisValid mirrors the genesis checks of `VerifyConfigChain`. -/
def genesisValid (obj : String) (g : Entry) (c : Config) : Bool :=
  g.object == obj && g.seqno == 0 && g.prev == "" && g.kind == ChangeType.genesis &&
    !c.participants.isEmpty && c.validate && verifySignatures g c

/--
verifyChain mirrors `VerifyConfigChain`: a valid genesis, then every later
entry verified against the configuration before it. Returns the final
configuration, which Go computes and discards.
-/
def verifyChain (obj : String) : List Entry → Option Config
  | [] => none
  | g :: rest =>
    match g.config with
    | none => none
    | some c =>
      if genesisValid obj g c then rest.foldlM (verifyChange obj) (c.withHead 0 g.hash) else none

/-! ## Configuration validity -/

/-- A valid nonempty configuration has an OWNER. -/
theorem Config.validate_owner {c : Config} (hv : c.validate = true) (hne : c.participants ≠ []) :
    ∃ p ∈ c.participants, p.role = Role.owner := by
  simp [Config.validate, Config.hasOwner, hne] at hv
  exact hv.2

/-- Moving the head of a nonempty configuration does not change its validity. -/
theorem Config.validate_withHead {c : Config} (hne : c.participants ≠ []) (seqno : Nat)
    (hash : String) : (c.withHead seqno hash).validate = c.validate := by
  have hne' : (c.withHead seqno hash).participants ≠ [] := hne
  simp only [Config.validate, List.isEmpty_eq_false_iff.mpr hne,
    List.isEmpty_eq_false_iff.mpr hne']
  rfl

/-- Configuration equality as Go compares it is transitive. -/
theorem Config.same_trans {a b c : Config} (hab : a.same b = true) (hbc : b.same c = true) :
    a.same c = true := by
  simp only [Config.same, Bool.and_eq_true, beq_iff_eq, List.isPerm_iff] at hab hbc ⊢
  obtain ⟨⟨hh, hs⟩, hp⟩ := hab
  obtain ⟨⟨hh', hs'⟩, hp'⟩ := hbc
  exact ⟨⟨hh.trans hh', hs.trans hs'⟩, hp.trans hp'⟩

/-- OWNER authority does not depend on participant order. -/
theorem isOwner_perm {a b : Config} (hp : a.participants.Perm b.participants) (peer : String) :
    isOwner a peer = isOwner b peer := by
  apply Bool.eq_iff_iff.mpr
  simp only [isOwner, List.any_eq_true]
  exact ⟨fun ⟨p, hm, h⟩ => ⟨p, hp.mem_iff.1 hm, h⟩,
    fun ⟨p, hm, h⟩ => ⟨p, hp.mem_iff.2 hm, h⟩⟩

/-! ## Single changes -/

/--
A verified change is the entry's configuration at the entry's head, linked to the
previous head at the next sequence number, signed with authority, and valid.
-/
theorem verifyChange_spec {obj : String} {cur : Config} {e : Entry} {next : Config}
    (h : verifyChange obj cur e = some next) :
    ∃ c, e.config = some c ∧ next = c.withHead e.seqno e.hash ∧ e.object = obj ∧
      e.prev = cur.hash ∧ expectedSeqno cur < seqnoLimit ∧ e.seqno = expectedSeqno cur ∧
      verifySignatures e cur = true ∧ next.validate = true := by
  obtain ⟨object, seqno, config, sigs, prev, kind, hash, merge⟩ := e
  cases config with
  | none => simp [verifyChange] at h
  | some c =>
    simp only [verifyChange] at h
    split at h
    · rename_i hok
      split at h
      · cases h
        exact ⟨c, rfl, rfl, hok⟩
      · contradiction
    · contradiction

/--
A verified change carries at least one signature, and every signature is valid
and, unless the change is a self-enrollment, from an OWNER of the previous
configuration.
-/
theorem verifyChange_authorized {obj : String} {cur next : Config} {e : Entry}
    (h : verifyChange obj cur e = some next) :
    e.sigs ≠ [] ∧ ∀ s ∈ e.sigs, s.valid = true ∧
      (e.kind ≠ ChangeType.selfEnrollPeer → isOwner cur s.signer = true) := by
  obtain ⟨_, _, _, _, _, _, _, hsig, _⟩ := verifyChange_spec h
  simp only [verifySignatures, Bool.and_eq_true, Bool.not_eq_true',
    List.isEmpty_eq_false_iff, List.all_eq_true] at hsig
  obtain ⟨⟨⟨hne, _⟩, _⟩, hall⟩ := hsig
  refine ⟨hne, fun s hs => ?_⟩
  obtain ⟨⟨_, hv⟩, hauth⟩ := hall s hs
  refine ⟨hv, fun hk => ?_⟩
  have hk' : (e.kind == ChangeType.selfEnrollPeer) = false := by simpa using hk
  simpa [signerAuthorized, hk'] using hauth

/-- A verified change leaves an OWNER in every nonempty configuration. -/
theorem verifyChange_owner {obj : String} {cur next : Config} {e : Entry}
    (h : verifyChange obj cur e = some next) (hne : next.participants ≠ []) :
    ∃ p ∈ next.participants, p.role = Role.owner := by
  obtain ⟨_, _, _, _, _, _, _, _, hv⟩ := verifyChange_spec h
  exact Config.validate_owner hv hne

/--
A change at or below the held sequence number is rejected, so a held head
cannot be forked or rolled back.
-/
theorem verifyChange_stale {obj : String} {cur : Config} {e : Entry} (hh : cur.hash ≠ "")
    (hs : e.seqno ≤ cur.seqno) : verifyChange obj cur e = none := by
  cases h : verifyChange obj cur e with
  | none => rfl
  | some next =>
    obtain ⟨_, _, _, _, _, _, hseq, _, _⟩ := verifyChange_spec h
    simp [expectedSeqno, hh] at hseq
    omega

/-! ## Suffix verification -/

/-- Configuration equality as Go compares it is symmetric. -/
theorem Config.same_symm {a b : Config} (h : a.same b = true) : b.same a = true := by
  simp only [Config.same, Bool.and_eq_true, beq_iff_eq, List.isPerm_iff] at h ⊢
  obtain ⟨⟨hh, hs⟩, hp⟩ := h
  exact ⟨⟨hh.symm, hs.symm⟩, hp.symm⟩

/-- Validity does not depend on participant order. -/
theorem Config.validate_perm {a b : Config} (hp : a.participants.Perm b.participants)
    (hh : a.hash = b.hash) : a.validate = b.validate := by
  have hall : ∀ f : Participant → Bool, a.participants.all f = b.participants.all f := by
    intro f
    apply Bool.eq_iff_iff.mpr
    simp only [List.all_eq_true]
    exact ⟨fun h x hx => h x (hp.mem_iff.2 hx), fun h x hx => h x (hp.mem_iff.1 hx)⟩
  have hany : ∀ f : Participant → Bool, a.participants.any f = b.participants.any f := by
    intro f
    apply Bool.eq_iff_iff.mpr
    simp only [List.any_eq_true]
    exact ⟨fun ⟨x, hx, h⟩ => ⟨x, hp.mem_iff.1 hx, h⟩, fun ⟨x, hx, h⟩ => ⟨x, hp.mem_iff.2 hx, h⟩⟩
  have hempty : a.participants.isEmpty = b.participants.isEmpty := by
    apply Bool.eq_iff_iff.mpr
    simp only [List.isEmpty_iff]
    exact ⟨fun h => List.perm_nil.1 (h ▸ hp).symm, fun h => List.perm_nil.1 (h ▸ hp.symm).symm⟩
  simp only [Config.validate, Config.hasOwner, hempty, hh, hall, hany,
    (hp.map (·.peer)).nodup_iff]

/-- Merging a change keeps a permutation of the target a permutation. -/
theorem mergeParticipants_perm {t t' : List Participant} (hp : t.Perm t')
    (parent next : List Participant) :
    (mergeParticipants t parent next).Perm (mergeParticipants t' parent next) :=
  (hp.filter _).append_right _

/-- Absorbing a record depends on the merged configuration only up to Go's equality. -/
theorem absorb_same {a b : Config} (h : a.same b = true) (d : Delta) :
    (absorb a d).same (absorb b d) = true := by
  have h' := h
  simp only [Config.same, Bool.and_eq_true, beq_iff_eq, List.isPerm_iff] at h'
  obtain ⟨⟨hh, hs⟩, hp⟩ := h'
  have hmp := mergeParticipants_perm hp d.parent.participants d.next.participants
  have hown : isOwner a = isOwner b := funext (isOwner_perm hp)
  have hval := Config.validate_perm (a := { a with
    participants := mergeParticipants a.participants d.parent.participants d.next.participants })
    (b := { b with
      participants := mergeParticipants b.participants d.parent.participants d.next.participants })
    hmp hh
  unfold absorb
  simp only [hown, hval]
  split
  · simp only [Config.same, Bool.and_eq_true, beq_iff_eq, List.isPerm_iff]
    exact ⟨⟨hh, hs⟩, hmp⟩
  · exact h

/-- Absorbing a branch depends on the merged configuration only up to Go's equality. -/
theorem foldl_absorb_same {a b : Config} (h : a.same b = true) (ds : List Delta) :
    (ds.foldl absorb a).same (ds.foldl absorb b) = true := by
  induction ds generalizing a b with
  | nil => exact h
  | cons d ds ih => exact ih (absorb_same h d)

/-- Merge verification depends on the held configuration only up to Go's equality. -/
theorem mergeValid_congr {obj : String} {a b : Config} (h : a.same b = true) (kind : ChangeType)
    (next : Config) (m : Merge) : mergeValid obj a kind next m = mergeValid obj b kind next m := by
  cases m with
  | none => rfl
  | some fork branch =>
    simp only [mergeValid]
    congr 1
    split
    · rfl
    · split
      · apply Bool.eq_iff_iff.mpr
        exact ⟨fun hx => Config.same_trans hx (foldl_absorb_same h _),
          fun hx => Config.same_trans hx (Config.same_symm (foldl_absorb_same h _))⟩
      · rfl

/-- A suffix step depends on the configuration only up to Go's equality. -/
theorem suffixStep_congr {obj : String} {a b : Config} (h : a.same b = true) (e : Entry) :
    suffixStep obj a e = suffixStep obj b e := by
  have hm := mergeValid_congr (obj := obj) h
  simp only [Config.same, Bool.and_eq_true, beq_iff_eq, List.isPerm_iff] at h
  obtain ⟨⟨hh, hs⟩, hp⟩ := h
  unfold suffixStep
  split
  · rename_i hk
    have hk' : (e.kind == ChangeType.selfEnrollPeer) = false := by
      simp only [ChangeType.peerChanges, List.mem_cons, List.not_mem_nil] at hk
      simp only [beq_eq_false_iff_ne]
      intro heq
      simp [heq, ChangeType.selfEnrollPeer, ChangeType.addParticipant,
        ChangeType.removeParticipant, ChangeType.addInvite, ChangeType.revokeInvite,
        ChangeType.incrementInviteUses, ChangeType.transferOwnership, ChangeType.setRoster,
        ChangeType.setSequencer, ChangeType.setControl, ChangeType.merge] at hk
    obtain ⟨object, seqno, config, sigs, prev, kind, hash, merge⟩ := e
    simp only at hk'
    simp [verifyChange, verifySignatures, signerAuthorized, expectedSeqno, hh, hs, hk',
      isOwner_perm hp, hm]
  · rfl

/-- Folding a suffix in two parts equals folding it at once. -/
theorem applySuffix_append (obj : String) (a : Config) (xs ys : List Entry) :
    applySuffix obj a (xs ++ ys) = (applySuffix obj a xs).bind (applySuffix obj · ys) := by
  simp [applySuffix, List.foldlM_append]

/-- A nonempty suffix gives the same result from Go-equal checkpoints. -/
theorem applySuffix_congr {obj : String} {a b : Config} (h : a.same b = true) {ys : List Entry}
    (hne : ys ≠ []) : applySuffix obj a ys = applySuffix obj b ys := by
  cases ys with
  | nil => contradiction
  | cons e ys => simp [applySuffix, List.foldlM_cons, suffixStep_congr h]

/-- Follows says a suffix starts at the head of `cur`, so no merge fork replaces it. -/
def Follows (cur : Config) (es : List Entry) : Prop :=
  ∀ e ∈ es.head?, e.prev = cur.hash

/-- A suffix that follows the checkpoint starts there. -/
theorem suffixBase_follows {cur : Config} {es : List Entry} (h : Follows cur es) :
    suffixBase cur es = some cur := by
  cases es with
  | nil => rfl
  | cons e es => simp [suffixBase, h e (by simp)]

/-- Follows is the link test `verifySuffix` makes. -/
theorem Follows.all {cur : Config} {es : List Entry} (h : Follows cur es) :
    es.head?.all (·.prev == cur.hash) = true := by
  cases es with
  | nil => rfl
  | cons e es => simp [h e (by simp)]

/-- An empty suffix accepts only a candidate equal to the checkpoint. -/
theorem verifySuffix_nil {obj : String} {a c : Config} (h : verifySuffix obj a c [] = true) :
    a.same c = true := by
  simp [verifySuffix, suffixBase, applySuffix] at h
  exact h.2

/--
Suffix verification composes: a verified path from `a` to `b` and one from `b`
to `c`, each following its checkpoint, join into a verified path from `a` to
`c`.
-/
theorem verifySuffix_trans {obj : String} {a b c : Config} {xs ys : List Entry}
    (hx : Follows a xs) (hy : Follows b ys)
    (hab : verifySuffix obj a b xs = true) (hbc : verifySuffix obj b c ys = true) :
    verifySuffix obj a c (xs ++ ys) = true := by
  -- The joined suffix follows `a`: through `xs`, or through `ys` when `xs` is empty.
  have hxy : Follows a (xs ++ ys) := by
    cases xs with
    | cons x xs => simpa [Follows] using hx
    | nil =>
      have hsame := verifySuffix_nil hab
      simp only [Config.same, Bool.and_eq_true, beq_iff_eq] at hsame
      intro e he
      rw [hy e (by simpa using he), ← hsame.1.1]
  simp only [verifySuffix, suffixBase_follows hx, suffixBase_follows hy, suffixBase_follows hxy,
    hx.all, hy.all, hxy.all, Bool.true_or, Bool.true_and, Option.bind_some,
    Bool.and_eq_true] at hab hbc ⊢
  obtain ⟨hcheck, hab⟩ := hab
  obtain ⟨_, hbc⟩ := hbc
  refine ⟨hcheck, ?_⟩
  split at hab
  · rename_i b' hb'
    rw [applySuffix_append, hb', Option.bind_some]
    cases ys with
    | nil =>
      simp only [applySuffix, List.foldlM_nil] at hbc ⊢
      exact Config.same_trans hab hbc
    | cons e ys =>
      rw [applySuffix_congr hab (List.cons_ne_nil e ys)]
      exact hbc
  · contradiction

/-! ## Full chains -/

/-- Every configuration reached through verified changes from a valid start is valid. -/
theorem foldlM_verifyChange_valid {obj : String} {es : List Entry} :
    ∀ {init c : Config}, init.validate = true → es.foldlM (verifyChange obj) init = some c →
      c.validate = true := by
  induction es with
  | nil => intro init c hv h; simp at h; exact h ▸ hv
  | cons e es ih =>
    intro init c _ h
    simp only [List.foldlM_cons] at h
    cases hstep : verifyChange obj init e with
    | none => simp [hstep] at h
    | some next =>
      simp only [hstep, Option.bind_eq_bind, Option.bind_some] at h
      obtain ⟨_, _, _, _, _, _, _, _, hv⟩ := verifyChange_spec hstep
      exact ih hv h

/--
Each verified change advances a nonempty head by exactly one, given nonempty
entry hashes, which SHA-256 digests always are.
-/
theorem foldlM_verifyChange_seqno {obj : String} {es : List Entry} :
    ∀ {init c : Config}, init.hash ≠ "" → (∀ e ∈ es, e.hash ≠ "") →
      es.foldlM (verifyChange obj) init = some c →
      c.seqno = init.seqno + es.length ∧ c.hash ≠ "" := by
  induction es with
  | nil => intro init c hh _ h; simp at h; subst h; exact ⟨rfl, hh⟩
  | cons e es ih =>
    intro init c hh hes h
    simp only [List.foldlM_cons] at h
    cases hstep : verifyChange obj init e with
    | none => simp [hstep] at h
    | some next =>
      simp only [hstep, Option.bind_eq_bind, Option.bind_some] at h
      obtain ⟨_, _, hnext, _, _, _, hseq, _, _⟩ := verifyChange_spec hstep
      have hnh : next.hash ≠ "" := by
        simpa [hnext, Config.withHead] using hes e (List.mem_cons_self ..)
      obtain ⟨hs, hc⟩ := ih hnh (fun x hx => hes x (List.mem_cons_of_mem e hx)) h
      simp [expectedSeqno, hh] at hseq
      simp [hnext, Config.withHead] at hs
      simp only [List.length_cons]
      exact ⟨by omega, hc⟩

/-- A verified chain is a valid genesis followed by verified changes. -/
theorem verifyChain_spec {obj : String} {es : List Entry} {c : Config}
    (h : verifyChain obj es = some c) :
    ∃ g rest gc, es = g :: rest ∧ g.config = some gc ∧ genesisValid obj g gc = true ∧
      rest.foldlM (verifyChange obj) (gc.withHead 0 g.hash) = some c := by
  unfold verifyChain at h
  split at h
  · contradiction
  · rename_i g rest
    split at h
    · contradiction
    · rename_i gc hgc
      split at h
      · rename_i hg
        exact ⟨g, rest, gc, rfl, hgc, hg, h⟩
      · contradiction

/-- Every nonempty configuration a verified chain reaches has an OWNER. -/
theorem verifyChain_owner {obj : String} {es : List Entry} {c : Config}
    (h : verifyChain obj es = some c)
    (hne : c.participants ≠ []) : ∃ p ∈ c.participants, p.role = Role.owner := by
  obtain ⟨g, _, gc, _, _, hg, hfold⟩ := verifyChain_spec h
  simp only [genesisValid, Bool.and_eq_true, Bool.not_eq_true', List.isEmpty_eq_false_iff] at hg
  obtain ⟨⟨⟨_, hgne⟩, hgv⟩, _⟩ := hg
  have hv := foldlM_verifyChange_valid (by rwa [Config.validate_withHead hgne]) hfold
  exact Config.validate_owner hv hne

/--
A verified chain's head sequence number counts its entries after genesis,
given nonempty entry hashes.
-/
theorem verifyChain_seqno {obj : String} {es : List Entry} {c : Config}
    (h : verifyChain obj es = some c)
    (hes : ∀ e ∈ es, e.hash ≠ "") : c.seqno + 1 = es.length := by
  obtain ⟨g, rest, gc, rfl, _, _, hfold⟩ := verifyChain_spec h
  have hgh : (gc.withHead 0 g.hash).hash ≠ "" := hes g (List.mem_cons_self ..)
  obtain ⟨hs, _⟩ := foldlM_verifyChange_seqno hgh
    (fun e he => hes e (List.mem_cons_of_mem g he)) hfold
  simp [Config.withHead] at hs
  simp [hs]

/-! ## Self-enrollment -/

/--
Self-enrollment adds only its signer, with no more than the strongest role its
entity already holds; every other participant was already present.
-/
theorem validateSelfEnroll_added {cur next : Config} {signer : String}
    (h : validateSelfEnroll cur next signer = true) :
    ∃ a ∈ next.participants, a.peer = signer ∧ entityRole cur a.entity ≠ Role.unknown ∧
      a.role ≤ entityRole cur a.entity ∧
      ∀ p ∈ next.participants, p = a ∨ ∃ q ∈ cur.participants, q.peer = p.peer := by
  unfold validateSelfEnroll at h
  simp only [Bool.and_eq_true] at h
  obtain ⟨_, hadd⟩ := h
  split at hadd
  · rename_i a hsplit
    simp only [Bool.and_eq_true, beq_iff_eq, bne_iff_ne, ne_eq, decide_eq_true_eq] at hadd
    obtain ⟨⟨⟨hpeer, _⟩, hrole⟩, hle⟩ := hadd
    have ha : a ∈ next.participants.filter
        (fun p => !cur.participants.any (·.peer == p.peer)) := by
      rw [hsplit]; exact List.mem_singleton_self a
    refine ⟨a, (List.mem_filter.1 ha).1, hpeer, hrole, hle, fun p hp => ?_⟩
    by_cases hin : cur.participants.any (·.peer == p.peer) = true
    · simp only [List.any_eq_true, beq_iff_eq] at hin
      exact Or.inr hin
    · have hpf : p ∈ next.participants.filter
          (fun p => !cur.participants.any (·.peer == p.peer)) :=
        List.mem_filter.2 ⟨hp, by simpa using hin⟩
      rw [hsplit] at hpf
      exact Or.inl (List.mem_singleton.1 hpf)
  · contradiction

end Spacewave.SObject
