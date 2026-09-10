package provider_spacewave

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	api "github.com/s4wave/spacewave/core/provider/spacewave/api"
	"github.com/s4wave/spacewave/core/session"
)

func TestLinkSessionUsesAuthorizerAndReceivingIdentity(t *testing.T) {
	sourceKey, sourcePeer := generateTestKeypair(t)
	_, receivingPeer := generateTestKeypair(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/account/session/link" || r.Header.Get("X-Peer-ID") != sourcePeer.String() {
			t.Errorf("unexpected Session enrollment request: %s %s by %s", r.Method, r.URL.Path, r.Header.Get("X-Peer-ID"))
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		request := &api.RegisterSessionRequest{}
		if err := request.UnmarshalVT(data); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.GetSessionPeerId() != receivingPeer.String() || request.GetType() != session.SessionType_SESSION_TYPE_USER || request.GetLabel() != "Receiving client" {
			t.Error("enrollment did not preserve the selected receiving Session")
		}
		response, err := (&api.RegisterSessionResponse{AccountId: "account", PeerId: receivingPeer.String()}).MarshalVT()
		if err != nil {
			t.Error(err)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(response)
	}))
	defer server.Close()
	client := NewSessionClient(server.Client(), server.URL, DefaultSigningEnvPrefix, sourceKey, sourcePeer.String())
	if _, err := client.LinkSession(context.Background(), receivingPeer, "Receiving client"); err != nil {
		t.Fatal(err)
	}
}

func TestMountHandoffSessionVerifiesAccountBeforePersistence(t *testing.T) {
	key, _ := generateTestKeypair(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := (&api.AccountInfoResponse{AccountId: "other-account"}).MarshalVT()
		_, _ = w.Write(data)
	}))
	defer server.Close()
	account := NewTestProviderAccount(t, server.URL)
	_, err := account.p.MountHandoffSession(context.Background(), "selected-account", key, nil)
	if err == nil || !strings.Contains(err.Error(), "belongs to another account") {
		t.Fatalf("unexpected handoff validation result: %v", err)
	}
}
