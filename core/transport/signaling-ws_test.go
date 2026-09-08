package transport

import (
	"context"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	ws "github.com/aperturerobotics/go-websocket"
	"github.com/pkg/errors"
	api "github.com/s4wave/spacewave/core/provider/spacewave/api"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/s4wave/spacewave/net/signaling"
	"github.com/sirupsen/logrus"
)

// TestWSSignalPeerResolverWaitsForConnection verifies cancellation before signaling is ready.
func TestWSSignalPeerResolverWaitsForConnection(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	ctrl := &wsSignalingCtrl{ready: make(chan struct{})}
	resolver := &wsSignalPeerResolver{
		c:   ctrl,
		dir: signaling.NewSignalPeer("webrtc", peer.ID("local"), peer.ID("remote")),
	}
	if err := resolver.Resolve(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("resolve error = %v, want context cancellation while waiting for signaling", err)
	}
}

// TestWSSignalingControllerRefreshesTicketAfterConnectionExpires verifies that
// a disconnected WebSocket reconnects with a newly issued authorization ticket.
func TestWSSignalingControllerRefreshesTicketAfterConnectionExpires(t *testing.T) {
	// Generate the identity used to request tickets from the test cloud.
	ctx, cancel := context.WithTimeout(t.Context(), 25*time.Second)
	t.Cleanup(cancel)
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}

	// Issue unique tickets so a stale URL cannot authorize a later generation.
	var mtx sync.Mutex
	issuedTickets := make([]string, 0, 2)
	authorizedTickets := make([]string, 0, 2)
	validTicket := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Issue the typed cloud response used by the production ticket client.
		if r.Method == http.MethodPost && r.URL.Path == "/api/signal/ticket" {
			mtx.Lock()
			ticket := "ticket-" + strconv.Itoa(len(issuedTickets)+1)
			issuedTickets = append(issuedTickets, ticket)
			validTicket = ticket
			mtx.Unlock()
			data, err := (&api.SignalTicketResponse{Token: ticket}).MarshalVT()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			if _, err := w.Write(data); err != nil {
				t.Error(err)
			}
			return
		}

		// Reject reuse of the ticket whose first connection has expired.
		mtx.Lock()
		ticket := r.URL.Query().Get("tk")
		valid := ticket != "" && ticket == validTicket
		if valid {
			validTicket = ""
		}
		mtx.Unlock()
		if !valid {
			http.Error(w, "expired signaling ticket", http.StatusUnauthorized)
			return
		}

		// Drop the first accepted socket, then end the check at the second handshake.
		conn, err := ws.Accept(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.CloseNow()
		mtx.Lock()
		authorizedTickets = append(authorizedTickets, ticket)
		complete := len(authorizedTickets) == 2
		mtx.Unlock()
		if complete {
			cancel()
		}
	}))
	t.Cleanup(server.Close)

	// Exercise the real ticket client, URL builder, WebSocket dial, and retry loop.
	ctrl := &wsSignalingCtrl{
		le:         logrus.NewEntry(logrus.New()),
		ready:      make(chan struct{}),
		priv:       priv,
		sigID:      "webrtc",
		pid:        pid,
		retryDelay: time.Millisecond,
		url: func(ctx context.Context) (string, error) {
			ticket, err := acquireSignalTicket(ctx, server.URL, priv, pid, "")
			if err != nil {
				return "", err
			}
			return signalWebSocketURL(server.URL, ticket)
		},
	}
	if err := ctrl.Execute(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("execute error = %v, want context cancellation", err)
	}

	// Snapshot endpoint state before checking the generation sequence.
	mtx.Lock()
	gotIssued := slices.Clone(issuedTickets)
	gotAuthorized := slices.Clone(authorizedTickets)
	mtx.Unlock()
	if want := []string{"ticket-1", "ticket-2"}; !slices.Equal(gotIssued, want) {
		t.Fatalf("issued tickets = %v, want %v", gotIssued, want)
	}
	if want := []string{"ticket-1", "ticket-2"}; !slices.Equal(gotAuthorized, want) {
		t.Fatalf("authorized tickets = %v, want %v", gotAuthorized, want)
	}
}
