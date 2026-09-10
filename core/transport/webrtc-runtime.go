//go:build !tinygo

package transport

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	transport_controller "github.com/s4wave/spacewave/net/transport/controller"
	"github.com/s4wave/spacewave/net/transport/webrtc"
	"github.com/s4wave/spacewave/net/transport/websocket"
	"github.com/sirupsen/logrus"
)

// sessionTransportFactories supplies the native session transport controllers.
func sessionTransportFactories(b bus.Bus) []controller.Factory {
	return []controller.Factory{
		websocket.NewFactory(b),
		webrtc.NewFactory(b),
	}
}

// startWebRTCControllers starts authenticated signaling and its WebRTC transport.
func (t *SessionTransport) startWebRTCControllers(
	ctx context.Context,
	le *logrus.Entry,
	b bus.Bus,
) (*transport_controller.Controller, func(), error) {
	if t.signalingURL == "" {
		return nil, nil, nil
	}

	// Keep the initial ticket request synchronous so startup reports its
	// existing signaling error at the webrtc-controllers stage.
	ticket, err := acquireSignalTicket(ctx, t.signalingURL, t.sessionKey, t.peerID, t.signingEnvPfx)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, nil, ctxErr
		}
		return nil, nil, err
	}
	if _, err := signalWebSocketURL(t.signalingURL, ticket); err != nil {
		return nil, nil, err
	}

	// Refresh the short-lived ticket before each signaling connection attempt.
	le.Debug("connecting to signaling")
	sigCtrl := newWSSignalingCtrl(le, b, func(ctx context.Context) (string, error) {
		ticket, err := acquireSignalTicket(ctx, t.signalingURL, t.sessionKey, t.peerID, t.signingEnvPfx)
		if err != nil {
			return "", err
		}
		return signalWebSocketURL(t.signalingURL, ticket)
	}, t.sessionKey, "webrtc", t.peerID)
	if _, err := b.AddController(ctx, sigCtrl, nil); err != nil {
		return nil, nil, err
	}

	// Keep the transport reference alive for the session transport lifetime.
	rtcCtrl, _, rtcRef, err := loader.WaitExecControllerRunningTyped[*transport_controller.Controller](
		ctx, b,
		resolver.NewLoadControllerWithConfig(&webrtc.Config{
			SignalingId: "webrtc",
			WebRtc: &webrtc.WebRtcConfig{
				IceServers: []*webrtc.IceServerConfig{
					{Urls: []string{"stun:stun.l.google.com:19302"}},
				},
			},
			AllPeers: true,
		}),
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	le.Debug("signaling and webrtc controllers started")
	return rtcCtrl, rtcRef.Release, nil
}
