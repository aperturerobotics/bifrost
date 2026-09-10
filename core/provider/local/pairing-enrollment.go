package provider_local

import (
	"context"
	"time"

	"github.com/aperturerobotics/util/ulid"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/provider"
	"github.com/s4wave/spacewave/core/session"
	"github.com/s4wave/spacewave/db/kvtx"
	"github.com/s4wave/spacewave/net/peer"
	stream_packet "github.com/s4wave/spacewave/net/stream/packet"
)

// pairingEnrollment retains the selected account and receiving Session while
// the authenticated stream authorizes and persists their attachment.
type pairingEnrollment struct {
	offer     *PairingAccount
	identity  *PairingIdentity
	replica   *ProviderAccount
	session   session.Session
	release   func()
	proof     string
	offering  bool
	source    peer.ID
	receiving peer.ID
}

// preparePairingEnrollment exchanges identities without granting account access.
func (a *ProviderAccount) preparePairingEnrollment(ctx context.Context, stream *stream_packet.Session, offering bool, localPeer, remotePeer peer.ID) (*pairingEnrollment, error) {
	// The offering client selects the existing account and creates a fresh operation.
	if offering {
		settings, err := a.GetAccountSettingsRef(ctx)
		if err != nil {
			return nil, err
		}
		storagePeer, err := a.vol.GetPeer(ctx, true)
		if err != nil {
			return nil, err
		}
		settingsSO, releaseSettings, err := a.MountSharedObject(ctx, settings, nil)
		if err != nil {
			return nil, err
		}
		defer releaseSettings()
		snapshot, err := settingsSO.GetSharedObjectState(ctx)
		if err != nil {
			return nil, err
		}
		accountSettings, _, err := decodeAccountSettingsSnapshot(ctx, snapshot)
		if err != nil {
			return nil, err
		}
		name := accountSettings.GetDisplayName()
		if name == "" {
			name = "Local account"
		}
		offer := &PairingAccount{AccountId: a.GetAccountID(), SettingsId: settings.GetProviderResourceRef().GetId(), OperationId: ulid.NewULID(), StoragePeerId: storagePeer.GetPeerID().String(), DisplayName: name}
		if err := stream.SendMsg(&PairingFrame{Body: &PairingFrame_Account{Account: offer}}); err != nil {
			return nil, err
		}
		frame, err := receivePairingFrame(stream)
		if err != nil {
			return nil, err
		}
		identity := frame.GetIdentity()
		if err := validatePairingIdentity(offer, identity, localPeer, remotePeer); err != nil {
			return nil, err
		}
		proof, err := pairingApprovalContext(offer, identity, localPeer, remotePeer)
		if err != nil {
			return nil, err
		}
		return &pairingEnrollment{offer: offer, identity: identity, offering: true, proof: proof, source: localPeer, receiving: remotePeer, release: func() {}}, nil
	}

	// Open the source account in this machine's store while preserving other accounts.
	frame, err := receivePairingFrame(stream)
	if err != nil {
		return nil, err
	}
	offer := frame.GetAccount()
	if offer.GetAccountId() == "" || offer.GetSettingsId() == "" || offer.GetOperationId() == "" {
		return nil, errors.New("pairing source did not offer an account")
	}
	if _, _, err := peer.ParsePeerIDWithPubKey(offer.GetStoragePeerId()); err != nil {
		return nil, errors.Wrap(err, "invalid source storage identity")
	}
	account, releaseAccount, err := a.t.p.AccessProviderAccount(ctx, offer.GetAccountId(), nil)
	if err != nil {
		return nil, err
	}
	replica := account.(*ProviderAccount)
	ref, err := replica.preparePairingSession(ctx)
	if err != nil {
		releaseAccount()
		return nil, err
	}
	sess, releaseSession, err := replica.MountSession(ctx, ref, nil)
	if err != nil {
		releaseAccount()
		return nil, err
	}
	release := func() {
		releaseSession()
		releaseAccount()
	}

	// Prove the new Session and volume keys before either user approves the transfer.
	identity, err := replica.buildPairingIdentity(ctx, offer, sess, remotePeer, localPeer)
	if err != nil {
		release()
		return nil, err
	}
	proof, err := pairingApprovalContext(offer, identity, remotePeer, localPeer)
	if err != nil {
		release()
		return nil, err
	}
	if err := stream.SendMsg(&PairingFrame{Body: &PairingFrame_Identity{Identity: identity}}); err != nil {
		release()
		return nil, err
	}
	return &pairingEnrollment{offer: offer, identity: identity, replica: replica, session: sess, proof: proof, source: remotePeer, receiving: localPeer, release: release}, nil
}

// receivePairingFrame returns a protocol frame or the remote operation failure.
func receivePairingFrame(stream *stream_packet.Session) (*PairingFrame, error) {
	frame := &PairingFrame{}
	if err := stream.RecvMsg(frame); err != nil {
		return nil, err
	}
	if message := frame.GetError(); message != "" {
		return nil, errors.New(message)
	}
	return frame, nil
}

// preparePairingSession persists one receiving Session reference per account
// replica. Repeated pairing after interruption reuses its own durable credential.
func (a *ProviderAccount) preparePairingSession(ctx context.Context) (*session.SessionRef, error) {
	// An account already attached on this machine keeps its existing Session.
	controller, releaseController, err := session.ExLookupSessionController(ctx, a.t.p.b, "", false, nil)
	if err != nil {
		return nil, err
	}
	defer releaseController.Release()
	entries, err := controller.ListSessions(ctx)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		ref := entry.GetSessionRef()
		providerRef := ref.GetProviderResourceRef()
		if providerRef.GetProviderId() == a.GetProviderID() && providerRef.GetProviderAccountId() == a.GetAccountID() {
			return ref, nil
		}
	}

	// Serialize the binding with account-local mutations.
	release, err := a.mtx.Lock(ctx)
	if err != nil {
		return nil, err
	}
	defer release()
	store, releaseStore, err := a.buildSoObjectStore(ctx)
	if err != nil {
		return nil, err
	}
	defer releaseStore()
	key := SobjectBindingKey("pairing-session")
	var ref *session.SessionRef
	err = kvtx.RunTransaction(ctx, false, func(ctx context.Context) (kvtx.Tx, error) {
		return store.NewTransaction(ctx, false)
	}, func(ctx context.Context, tx kvtx.Tx) error {
		data, found, err := tx.Get(ctx, key)
		if err != nil || !found {
			return err
		}
		ref = &session.SessionRef{}
		if err := ref.UnmarshalVT(data); err != nil {
			return err
		}
		if err := ref.Validate(); err != nil {
			return err
		}
		if ref.GetProviderResourceRef().GetProviderAccountId() != a.GetAccountID() || ref.GetProviderResourceRef().GetProviderId() != a.GetProviderID() {
			return errors.New("pairing Session binding belongs to another account")
		}
		return nil
	})
	if err != nil || ref != nil {
		return ref, err
	}

	// Reserve the resumable reference before mounting once to persist its key.
	ref = &session.SessionRef{ProviderResourceRef: &provider.ProviderResourceRef{
		Id: ulid.NewULID(), ProviderId: a.GetProviderID(), ProviderAccountId: a.GetAccountID(),
	}}
	data, err := ref.MarshalVT()
	if err != nil {
		return nil, err
	}
	err = kvtx.RunTransaction(ctx, true, func(ctx context.Context) (kvtx.Tx, error) {
		return store.NewTransaction(ctx, true)
	}, func(ctx context.Context, tx kvtx.Tx) error {
		return tx.Set(ctx, key, data)
	})
	return ref, err
}

// completePairingEnrollment transfers authorized state after bilateral approval
// and waits for the receiving machine's durable Session registration.
func (a *ProviderAccount) completePairingEnrollment(ctx context.Context, stream *stream_packet.Session, enrollment *pairingEnrollment) error {
	// The source grants each actual receiving identity and streams bounded checkpoints.
	if enrollment.offering {
		release, err := a.replicaAuth.Lock(ctx)
		if err != nil {
			return err
		}
		defer release()
		if err := a.registerPairingReplicas(ctx, enrollment); err != nil {
			return err
		}
		for _, entry := range a.soListCtr.GetValue().GetSharedObjects() {
			object, err := a.enrollPairingObject(ctx, entry, enrollment.identity)
			if err != nil {
				return errors.Wrap(err, "enroll SharedObject "+entry.GetRef().GetProviderResourceRef().GetId())
			}
			if err := stream.SendMsg(&PairingFrame{Body: &PairingFrame_Object{Object: object}}); err != nil {
				return err
			}
		}
		if err := stream.SendMsg(&PairingFrame{Body: &PairingFrame_Complete{Complete: true}}); err != nil {
			return err
		}
		frame, err := receivePairingFrame(stream)
		if err != nil {
			return err
		}
		if !frame.GetComplete() {
			return errors.New("receiving client did not acknowledge account enrollment")
		}
		remotePeer, _, err := peer.ParsePeerIDWithPubKey(enrollment.identity.GetSessionProof().GetResponderPeerId())
		if err != nil {
			return err
		}
		return a.retainPairedAccountPeer(ctx, remotePeer)
	}

	// Bind the canonical settings identity before importing the replica's catalog.
	replica := enrollment.replica
	if err := replica.bindPairingSettings(ctx, enrollment.offer); err != nil {
		return err
	}
	settingsReceived := false
	for {
		frame, err := receivePairingFrame(stream)
		if err != nil {
			return err
		}
		if frame.GetComplete() {
			break
		}
		object := frame.GetObject()
		if err := replica.installPairingObject(ctx, enrollment.offer, object, enrollment.source); err != nil {
			return err
		}
		if object.GetEntry().GetRef().GetProviderResourceRef().GetId() == enrollment.offer.GetSettingsId() {
			settingsReceived = true
		}
	}
	if !settingsReceived {
		return errors.New("source did not provide the canonical account settings")
	}

	// Publish the durable account attachment through the ordinary Session controller.
	if err := replica.EnsureConfiguredSessionTransport(ctx, enrollment.session.GetPrivKey()); err != nil {
		return err
	}
	if err := replica.retainPairedAccountPeer(ctx, enrollment.source); err != nil {
		return err
	}
	controller, releaseController, err := session.ExLookupSessionController(ctx, a.t.p.b, "", false, nil)
	if err != nil {
		return err
	}
	defer releaseController.Release()
	ref := enrollment.session.GetSessionRef()
	metadata := &session.SessionMetadata{
		ProviderId: "local", ProviderDisplayName: "Local", ProviderAccountId: replica.GetAccountID(), CreatedAt: time.Now().UnixMilli(),
	}
	entries, err := controller.ListSessions(ctx)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.GetSessionRef().EqualVT(ref) {
			metadata = nil
			break
		}
	}
	if _, err := controller.RegisterSession(ctx, ref, metadata); err != nil {
		return err
	}
	return stream.SendMsg(&PairingFrame{Body: &PairingFrame_Complete{Complete: true}})
}

// retainPairedAccountPeer persists reconnect demand through the account lifecycle.
func (a *ProviderAccount) retainPairedAccountPeer(ctx context.Context, remotePeer peer.ID) error {
	transport := a.GetSessionTransport()
	if transport == nil {
		return ErrNoSessionTransport
	}
	if err := a.StartPersistentP2PSync(ctx, transport); err != nil {
		return err
	}
	return a.RetainP2PPeer(ctx, remotePeer)
}
