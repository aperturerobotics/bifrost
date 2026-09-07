package provider_spacewave

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/aperturerobotics/util/ccontainer"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/sirupsen/logrus"
)

// TestCloudSOHostUploadsBeforeOperation keeps referenced blocks available when
// the cloud receives an operation, without waiting for background sync.
func TestCloudSOHostUploadsBeforeOperation(t *testing.T) {
	for _, tc := range []struct {
		name       string
		failUpload bool
	}{
		{name: "uploaded"},
		{name: "upload-failed", failUpload: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Only the cloud HTTP boundary is substituted; the operation owner and
			// signed operation use their production implementations.
			var uploaded, posted atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				posted.Store(true)
				if !uploaded.Load() {
					t.Error("operation reached the cloud before its blocks were uploaded")
				}
				w.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(server.Close)
			key, peerID := generateTestKeypair(t)
			client := NewSessionClient(server.Client(), server.URL, DefaultSigningEnvPrefix, key, peerID.String())
			client.executeWriteTicketAudience = func(ctx context.Context, resourceID string, audience writeTicketAudience, submit func(string) error) error {
				return submit("test-ticket")
			}
			uploadErr := errors.New("upload unavailable")
			host := &cloudSOHost{
				le:       logrus.New().WithField("test", t.Name()),
				client:   client,
				soID:     testSharedObjectID,
				stateCtr: ccontainer.NewCContainer(&sobject.SOState{}),
				forceBlockSync: func(context.Context) error {
					if tc.failUpload {
						return uploadErr
					}
					uploaded.Store(true)
					return nil
				},
			}

			// Failed uploads must not publish an operation whose inputs are absent.
			err := host.QueueOperation(t.Context(), peerID, func(nonce uint64) (*sobject.SOOperation, error) {
				return sobject.BuildSOOperation(testSharedObjectID, key, []byte("operation"), nonce, sobject.NewSOOperationLocalID())
			})
			if tc.failUpload {
				if !errors.Is(err, uploadErr) || posted.Load() {
					t.Fatalf("failed upload submitted operation: err=%v posted=%v", err, posted.Load())
				}
			} else if err != nil || !uploaded.Load() || !posted.Load() {
				t.Fatalf("operation did not upload then submit: err=%v uploaded=%v posted=%v", err, uploaded.Load(), posted.Load())
			}
		})
	}
}
