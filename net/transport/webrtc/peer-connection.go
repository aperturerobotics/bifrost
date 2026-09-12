package webrtc

import (
	"context"
	"errors"

	pion "github.com/pion/webrtc/v4"
)

// newPeerConnection obtains fresh credentials without changing established links.
func (w *WebRTC) newPeerConnection(ctx context.Context) (*pion.PeerConnection, error) {
	// Keep static configuration when no application credential source is present.
	config := w.webrtcConf
	if w.b != nil {
		provider, release, err := ExLookupICEProvider(ctx, w.b, w.peerID)
		if err != nil {
			return nil, err
		}
		if release != nil {
			defer release()
		}
		if provider != nil {
			fresh, err := provider.ICEConfig(ctx)
			if err != nil {
				return nil, err
			}
			if fresh == nil {
				return nil, errors.New("ICE provider returned no configuration")
			}
			config = fresh.ToWebRtcConfiguration()
		}
	}

	// Pion retains the selected configuration for this connection's lifetime.
	return w.webrtcApi.NewPeerConnection(*config)
}
