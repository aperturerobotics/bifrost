package provider_spacewave

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/pkg/errors"
)

// TestCloudMigrationFlushBeforeTransfer keeps pending writes under their source
// authority until the cloud has accepted them, including a failed flush.
func TestCloudMigrationFlushBeforeTransfer(t *testing.T) {
	for _, fail := range []bool{false, true} {
		name := "published"
		if fail {
			name = "flush-failed"
		}
		t.Run(name, func(t *testing.T) {
			var flushed, transferred atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/resource/transfer" || r.Method != http.MethodPost {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				if !flushed.Load() {
					t.Error("resource authority changed before pending writes were published")
				}
				transferred.Store(true)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			source := NewTestProviderAccount(t, server.URL)
			destination := NewTestProviderAccount(t, server.URL)
			flushErr := errors.New("cloud upload interrupted")
			object := &SharedObject{tkr: &sobjectTracker{id: "pending-space"}, blkStore: &BlockStore{forceSync: func(context.Context) error {
				if fail {
					return flushErr
				}
				flushed.Store(true)
				return nil
			}}}
			err := destination.ImportMigrationObject(t.Context(), source, object, nil, nil)
			if fail {
				if !errors.Is(err, flushErr) || transferred.Load() {
					t.Fatalf("failed flush changed resource authority: transferred=%v err=%v", transferred.Load(), err)
				}
			} else if err != nil || !transferred.Load() {
				t.Fatalf("published resource did not transfer: transferred=%v err=%v", transferred.Load(), err)
			}
		})
	}
}
