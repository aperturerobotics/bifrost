package provider_local

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"slices"

	"github.com/pkg/errors"
	account_settings "github.com/s4wave/spacewave/core/account/settings"
	"github.com/s4wave/spacewave/core/session"
	"github.com/s4wave/spacewave/core/sobject"
	sobject_invite "github.com/s4wave/spacewave/core/sobject/invite"
	"github.com/s4wave/spacewave/db/kvtx"
	"github.com/s4wave/spacewave/net/peer"
)

// buildPairingIdentity binds this replica's independently generated Session and
// storage keys to the offered account, operation, and authenticated transport.
func (a *ProviderAccount) buildPairingIdentity(ctx context.Context, offer *PairingAccount, sess session.Session, sourcePeer, receivingPeer peer.ID) (*PairingIdentity, error) {
	// Both proofs cover the exact receiving Session reference.
	ref := sess.GetSessionRef()
	proofContext, err := pairingIdentityContext(offer, ref, sourcePeer, receivingPeer)
	if err != nil {
		return nil, err
	}
	sessionProof, err := sobject_invite.BuildJoinResponse(proofContext, sess.GetPrivKey())
	if err != nil {
		return nil, err
	}

	// SharedObject mounts use the volume signer, independently of the Session.
	storagePeer, err := a.vol.GetPeer(ctx, true)
	if err != nil {
		return nil, err
	}
	storageKey, err := storagePeer.GetPrivKey(ctx)
	if err != nil {
		return nil, err
	}
	storageProof, err := sobject_invite.BuildJoinResponse(proofContext, storageKey)
	if err != nil {
		return nil, err
	}
	return &PairingIdentity{SessionRef: ref, SessionProof: sessionProof, StorageProof: storageProof}, nil
}

// pairingIdentityContext prevents proofs from authorizing another account,
// Session reference, transport connection, or pairing attempt.
func pairingIdentityContext(offer *PairingAccount, ref *session.SessionRef, sourcePeer, receivingPeer peer.ID) (string, error) {
	if offer.GetAccountId() == "" || offer.GetSettingsId() == "" || offer.GetOperationId() == "" {
		return "", errors.New("pairing account identity is incomplete")
	}
	if err := ref.Validate(); err != nil {
		return "", err
	}
	if ref.GetProviderResourceRef().GetProviderAccountId() != offer.GetAccountId() || ref.GetProviderResourceRef().GetProviderId() != "local" {
		return "", errors.New("receiving Session does not attach to the offered account")
	}
	if sourcePeer == "" || receivingPeer == "" || sourcePeer == receivingPeer {
		return "", errors.New("pairing requires distinct authenticated transport peers")
	}

	// Protobuf length-delimited fields keep the two serialized records unambiguous.
	data, err := (&PairingFrame{Body: &PairingFrame_Account{Account: offer}}).MarshalVT()
	if err != nil {
		return "", err
	}
	identity, err := (&PairingIdentity{SessionRef: ref}).MarshalVT()
	if err != nil {
		return "", err
	}
	digest := sha256.New()
	digest.Write(data)
	digest.Write(identity)
	digest.Write([]byte(sourcePeer.String() + "/" + receivingPeer.String()))
	return "account pairing/" + hex.EncodeToString(digest.Sum(nil)), nil
}

// validatePairingIdentity checks both proofs before any grant is changed.
func validatePairingIdentity(offer *PairingAccount, identity *PairingIdentity, sourcePeer, receivingPeer peer.ID) error {
	proofContext, err := pairingIdentityContext(offer, identity.GetSessionRef(), sourcePeer, receivingPeer)
	if err != nil {
		return err
	}
	for _, proof := range []*sobject.SOJoinResponse{identity.GetSessionProof(), identity.GetStorageProof()} {
		if proof.GetInviteId() != proofContext {
			return errors.New("pairing identity proof does not match the approved operation")
		}
		if _, _, err := sobject_invite.ValidateJoinResponse(proof); err != nil {
			return err
		}
	}
	return nil
}

// pairingApprovalContext binds bilateral approval to both complete signed proofs.
func pairingApprovalContext(offer *PairingAccount, identity *PairingIdentity, sourcePeer, receivingPeer peer.ID) (string, error) {
	proofContext, err := pairingIdentityContext(offer, identity.GetSessionRef(), sourcePeer, receivingPeer)
	if err != nil {
		return "", err
	}
	data, err := identity.MarshalVT()
	if err != nil {
		return "", err
	}
	digest := sha256.New()
	digest.Write([]byte(proofContext))
	digest.Write(data)
	return hex.EncodeToString(digest.Sum(nil)), nil
}

// enrollPairingObject grants the approved receiving identities and exports the
// resulting checkpoint. Its caller must have completed bilateral approval.
func (a *ProviderAccount) enrollPairingObject(ctx context.Context, entry *sobject.SharedObjectListEntry, identity *PairingIdentity) (*PairingSharedObject, error) {
	// Initial enrollment requires signed possession of both independent keys.
	for _, proof := range []*sobject.SOJoinResponse{identity.GetSessionProof(), identity.GetStorageProof()} {
		if _, _, err := sobject_invite.ValidateJoinResponse(proof); err != nil {
			return nil, err
		}
	}
	return a.enrollAccountMemberObject(ctx, entry, &account_settings.AccountSession{
		PeerId: identity.GetSessionProof().GetResponderPeerId(), StoragePeerId: identity.GetStorageProof().GetResponderPeerId(),
	})
}

// installPairingObject accepts a checkpoint only for the offered account. The
// authenticated enrollment exchange supplies its trust; the catalog does not.
func (a *ProviderAccount) installPairingObject(ctx context.Context, offer *PairingAccount, object *PairingSharedObject, sourcePeer peer.ID) error {
	entry := object.GetEntry()
	ref := entry.GetRef()
	if err := ref.Validate(); err != nil {
		return err
	}
	providerRef := ref.GetProviderResourceRef()
	if offer.GetAccountId() != a.GetAccountID() || providerRef.GetProviderAccountId() != a.GetAccountID() || providerRef.GetProviderId() != a.GetProviderID() {
		return errors.New("pairing object belongs to another account")
	}
	if object.GetState() == nil || entry.GetMeta().GetBodyType() == "" {
		return errors.New("pairing object checkpoint is incomplete")
	}
	return a.mountEnrolledSO(ctx, providerRef.GetId(), entry.GetMeta(), entry.GetSource(), object.GetState(), sourcePeer)
}

// bindPairingSettings replaces only the empty settings object created while
// opening a fresh replica. Existing account data prevents an identity change.
func (a *ProviderAccount) bindPairingSettings(ctx context.Context, offer *PairingAccount) error {
	// Serialize the persisted binding and local list update with account mutations.
	release, err := a.mtx.Lock(ctx)
	if err != nil {
		return err
	}
	defer release()
	current, err := a.GetAccountSettingsRef(ctx)
	if err != nil {
		return err
	}
	if current.GetProviderResourceRef().GetId() == offer.GetSettingsId() {
		return nil
	}
	list := a.soListCtr.GetValue()
	if len(list.GetSharedObjects()) != 1 || list.GetSharedObjects()[0].GetRef().GetProviderResourceRef().GetId() != current.GetProviderResourceRef().GetId() {
		return errors.New("existing account has a different settings identity; merge or repair is required")
	}

	// A fresh account's default object has no user operations or state to preserve.
	so, releaseSO, err := a.MountSharedObject(ctx, current, nil)
	if err != nil {
		return err
	}
	defer releaseSO()
	snapshot, err := so.GetSharedObjectState(ctx)
	if err != nil {
		return err
	}
	root, err := snapshot.GetRootInner(ctx)
	if err != nil {
		return err
	}
	if len(root.GetStateData()) != 0 {
		return errors.New("existing account settings contain data; merge or repair is required")
	}

	// Keep the old object durable until the new binding is committed.
	ref := sobject.NewSharedObjectRef(a.GetProviderID(), a.GetAccountID(), offer.GetSettingsId(), SobjectBlockStoreID(offer.GetSettingsId()))
	next := list.CloneVT()
	next.SharedObjects = slices.DeleteFunc(next.SharedObjects, func(entry *sobject.SharedObjectListEntry) bool {
		return entry.GetRef().GetProviderResourceRef().GetId() == current.GetProviderResourceRef().GetId()
	})
	store, releaseStore, err := a.buildSoObjectStore(ctx)
	if err != nil {
		return err
	}
	defer releaseStore()
	binding, err := ref.MarshalVT()
	if err != nil {
		return err
	}
	listData, err := next.MarshalVT()
	if err != nil {
		return err
	}
	if err := kvtx.RunTransaction(ctx, true, func(ctx context.Context) (kvtx.Tx, error) {
		return store.NewTransaction(ctx, true)
	}, func(ctx context.Context, tx kvtx.Tx) error {
		if err := tx.Set(ctx, SobjectBindingKey(accountSettingsBindingPurpose), binding); err != nil {
			return err
		}
		return tx.Set(ctx, SobjectObjectStoreListKey(), listData)
	}); err != nil {
		return err
	}
	a.soListCtr.SetValue(next)
	a.accountSettingsProcessor.SetRoutine(a.runAccountSettingsProcessor)
	return nil
}
