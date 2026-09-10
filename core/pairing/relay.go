package pairing

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"math/big"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	alpha_nethttp "github.com/s4wave/spacewave/core/nethttp"
	api "github.com/s4wave/spacewave/core/provider/spacewave/api"
	"github.com/s4wave/spacewave/core/transport"
	"github.com/s4wave/spacewave/net/link"
	"github.com/s4wave/spacewave/net/peer"
)

// Relay supplies the configured code and signaling service as one signing scope.
type Relay struct {
	URL              string
	SigningEnvPrefix string
	Client           *http.Client
}

func (r Relay) client() *http.Client {
	if r.Client != nil {
		return r.Client
	}
	return http.DefaultClient
}

// GenerateCode publishes a code after the Session transport is ready to accept
// an authenticated peer. Local and cloud accounts use the same relay contract.
func (e *Engine) GenerateCode(ctx context.Context, relay Relay) (string, error) {
	e.Clear()
	st, err := e.transport(ctx, relay)
	if err != nil {
		return "", err
	}
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 8)
	for i := range code {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		code[i] = alphabet[index.Int64()]
	}
	body, err := (&api.PairingRequest{Code: string(code), PeerId: e.peerID.String()}).MarshalVT()
	if err != nil {
		return "", err
	}
	endpoint, err := url.JoinPath(relay.URL, "/api/pair")
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	if err := transport.SignHTTPRequest(req, body, e.key, e.peerID, relay.SigningEnvPrefix); err != nil {
		return "", err
	}
	resp, err := relay.client().Do(req)
	if err != nil {
		return "", errors.Wrap(err, "register pairing code")
	}
	defer alpha_nethttp.DrainAndCloseResponseBody(resp)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("pairing code registration failed: HTTP %d", resp.StatusCode)
	}
	parentCtx, active := e.begin(true, true, string(code), "", StatusCodeGenerated)
	go e.runSolicit(parentCtx, active, st)
	return string(code), nil
}

// CompleteCode resolves a code and retains the peer link through enrollment.
func (e *Engine) CompleteCode(ctx context.Context, relay Relay, code string, offerCurrent bool) (peer.ID, error) {
	e.Clear()
	st, err := e.transport(ctx, relay)
	if err != nil {
		return "", err
	}
	endpoint, err := url.JoinPath(relay.URL, "/api/pair", code)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	resp, err := relay.client().Do(req)
	if err != nil {
		return "", errors.Wrap(err, "resolve pairing code")
	}
	defer alpha_nethttp.DrainAndCloseResponseBody(resp)
	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("pairing code lookup failed: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	response := &api.PairingResponse{}
	if err := response.UnmarshalVT(body); err != nil {
		return "", err
	}
	remotePeer, _, err := peer.ParsePeerIDWithPubKey(response.GetPeerId())
	if err != nil {
		return "", err
	}
	parentCtx, active := e.begin(false, offerCurrent, "", remotePeer, StatusWaitingForPeer)
	go func() {
		_, release, err := link.EstablishLinkWithPeerEx(parentCtx, st.GetChildBus(), e.peerID, remotePeer, false)
		if err != nil {
			e.fail(active, StatusFailed, err)
			return
		}
		defer release()
		e.update(active, func(a *attempt) { a.snapshot.Status = StatusPeerConnected })
		e.runSolicit(parentCtx, active, st)
	}()
	return remotePeer, nil
}
