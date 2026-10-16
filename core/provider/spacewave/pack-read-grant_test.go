package provider_spacewave

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/s4wave/spacewave/db/packfile"
)

// TestReadGrantsResendOnReset checks that a grant read survives the server
// closing a reused connection without a response.
func TestReadGrantsResendOnReset(t *testing.T) {
	// Serve grant reads on a server that counts them.
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Count the request.
		requests++

		// Close the connection of the second request without answering.
		if requests == 2 {
			conn, _, err := http.NewResponseController(w).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			_ = conn.Close()
			return
		}

		// Answer one grant.
		data, err := (&packfile.ReadResponse{Grants: []*packfile.ReadGrant{{PackId: "pack"}}}).MarshalVT()
		if err != nil {
			t.Error(err)
			return
		}
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	// The second read reuses the first read's connection and is resent once
	// the server closes it.
	priv, pid := generateTestKeypair(t)
	cli := NewSessionClient(srv.Client(), srv.URL, DefaultSigningEnvPrefix, priv, pid.String())
	for range 2 {
		if _, err := cli.ReadGrants(t.Context(), "store", []string{"pack"}); err != nil {
			t.Fatal(err)
		}
	}
	if requests != 3 {
		t.Fatalf("expected 3 requests, got %d", requests)
	}
}
