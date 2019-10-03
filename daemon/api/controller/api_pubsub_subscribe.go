package bifrost_api_controller

import (
	"context"
	"time"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/pubsub"
	"github.com/aperturerobotics/bifrost/pubsub/grpc"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/pkg/errors"
)

var (
	// publishTimeout is the timeout for acquiring private key and publishing
	publishTimeout = time.Second * 5
)

// Subscribe subscribes to a pubsub channel.
func (a *API) Subscribe(serv pubsub_grpc.PubSubService_SubscribeServer) error {
	ctx := serv.Context()

	var channelID string
	var handlePeerID string
	var handlePeer peer.Peer
	var sub pubsub.Subscription
	var handlePeerRef directive.Reference
	defer func() {
		if handlePeerRef != nil {
			handlePeerRef.Release()
		}
	}()
	for {
		msg, err := serv.Recv()
		if err != nil {
			return err
		}
		if chid := msg.GetChannelId(); chid != "" {
			if channelID != "" {
				return errors.New("channel id cannot be specified twice")
			}
			channelID = chid
			// acquire channel
			av, subRef, err := bus.ExecOneOff(
				ctx,
				a.bus,
				pubsub.NewBuildChannelSubscription(channelID),
				nil,
			)
			if err != nil {
				return err
			}
			defer subRef.Release()
			val, ok := av.GetValue().(pubsub.BuildChannelSubscriptionValue)
			if !ok {
				return errors.New("build channel subscription returned invalid value")
			}
			sub = val
			err = serv.Send(&pubsub_grpc.SubscribeResponse{
				SubscriptionStatus: &pubsub_grpc.SubscriptionStatus{
					Subscribed: true,
				},
			})
			if err != nil {
				return err
			}
			defer val.AddHandler(func(m pubsub.Message) {
				go func() {
					_ = serv.Send(&pubsub_grpc.SubscribeResponse{
						IncomingMessage: &pubsub_grpc.IncomingMessage{
							FromPeerId:    m.GetFrom().Pretty(),
							Data:          m.GetData(),
							Authenticated: m.GetAuthenticated(),
						},
					})
				}()
			})()
		}
		if channelID == "" {
			return errors.New("channel id must be specified in first message")
		}

		pubReqData := msg.GetPublishRequest().GetData()
		if len(pubReqData) == 0 {
			continue
		}
		peerID := msg.GetPublishRequest().GetFromPeerId()
		if peerID == "" {
			return errors.New("publish request: peer id cannot be empty")
		}
		if handlePeerID != peerID {
			if handlePeerRef != nil {
				handlePeerRef.Release()
				handlePeerRef = nil
				// handlePeerID = ""
				handlePeer = nil
			}

			npid, err := peer.IDB58Decode(peerID)
			if err != nil {
				return err
			}
			pubCtx, pubCtxCancel := context.WithTimeout(ctx, publishTimeout)
			handlePeerID = peerID
			handlePeer, handlePeerRef, err = peer.GetPeerWithID(pubCtx, a.bus, npid)
			pubCtxCancel()
			if err != nil || handlePeer == nil {
				return errors.Errorf("peer not identified locally: %s", npid.Pretty())
			}
		}

		if err := sub.Publish(handlePeer.GetPrivKey(), pubReqData); err != nil {
			return err
		}
	}
}
