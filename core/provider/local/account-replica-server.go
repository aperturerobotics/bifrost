package provider_local

import (
	"context"

	"github.com/s4wave/spacewave/core/pairing"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/starpc/srpc"
	"github.com/aperturerobotics/util/routine"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/net/link"
	"github.com/s4wave/spacewave/net/protocol"
	stream_srpc_server "github.com/s4wave/spacewave/net/stream/srpc/server"
)

// accountReplicaProtocol carries checkpoint requests authorized by account membership.
const accountReplicaProtocol = protocol.ID("alpha/account-replica/1")

// FetchObject authorizes the authenticated Session against the current canonical
// account registry. Catalog knowledge alone never permits checkpoint enrollment.
func (a *ProviderAccount) FetchObject(ctx context.Context, request *AccountReplicaObjectRequest) (*pairing.SharedObject, error) {
	release, err := a.replicaAuth.Lock(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

	stream, err := link.MustGetMountedStreamContext(ctx)
	if err != nil {
		return nil, err
	}
	ref, err := a.GetAccountSettingsRef(ctx)
	if err != nil {
		return nil, err
	}
	if request.GetSettingsId() != ref.GetProviderResourceRef().GetId() {
		return nil, errors.New("replica request belongs to another account")
	}
	settings, err := a.readAccountSettings(ctx)
	if err != nil {
		return nil, err
	}
	member := settings.FindAccountSession(stream.GetPeerID().String())
	if member == nil || member.GetRevoked() {
		return nil, errors.New("replica Session is not an active account member")
	}
	entry := settings.FindCatalogEntry(request.GetObjectId())
	if entry == nil || entry.GetDeleted() {
		return nil, errors.New("requested object is absent from the account catalog")
	}
	return a.enrollAccountMemberObject(ctx, entry.GetEntry(), member)
}

// startAccountReplicaSync attaches service and reconciliation to the existing P2P
// generation. Its ordinary stop path joins reconciliation before releasing mounts.
func (a *ProviderAccount) startAccountReplicaSync(state *p2pSyncState) error {
	transport := state.sessionTransport
	server, err := stream_srpc_server.NewServer(
		transport.GetChildBus(), a.le,
		controller.NewInfo("alpha/account-replica", controller.MustParseVersion("0.0.1"), "account replica service"),
		[]stream_srpc_server.RegisterFn{func(mux srpc.Mux) error { return SRPCRegisterAccountReplicaService(mux, a) }},
		[]protocol.ID{accountReplicaProtocol}, []string{transport.GetPeerID().String()}, false,
	)
	if err != nil {
		return err
	}
	release, err := transport.GetChildBus().AddController(state.ctx, server, nil)
	if err != nil {
		return err
	}
	state.addRelease(release)
	reconcile := routine.NewRoutineContainerWithLogger(a.le.WithField("routine", "account-replica"), routine.WithRetry(providerBackoff))
	reconcile.SetRoutine(func(ctx context.Context) error { return a.runAccountReplicaSync(ctx, state) })
	state.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) { state.replicaSync = reconcile })
	reconcile.SetContext(state.ctx, false)
	return nil
}

var _ SRPCAccountReplicaServiceServer = (*ProviderAccount)(nil)
