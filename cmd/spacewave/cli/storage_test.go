//go:build !js

package spacewave_cli

import (
	"testing"

	s4wave_session "github.com/s4wave/spacewave/sdk/session"
)

// TestFormatSpaceStorage checks the Storage line of space info for each
// placement, with and without pending blocks or an upload error.
func TestFormatSpaceStorage(t *testing.T) {
	cases := []struct {
		storage *s4wave_session.WatchSpaceStorageResponse
		want    string
	}{
		{&s4wave_session.WatchSpaceStorageResponse{}, "account storage, on this device"},
		{
			&s4wave_session.WatchSpaceStorageResponse{PendingBlocks: 3, PendingBytes: 2048},
			"account storage, 3 blocks (2.0 KiB) waiting to upload",
		},
		{&s4wave_session.WatchSpaceStorageResponse{StorageBackendName: "r2"}, "r2, uploaded"},
		{
			&s4wave_session.WatchSpaceStorageResponse{StorageBackendName: "r2", PendingBlocks: 1, PendingBytes: 10, UploadError: "denied"},
			"r2, 1 block (10 B) waiting to upload: denied",
		},
		{
			&s4wave_session.WatchSpaceStorageResponse{StorageBackendName: "r2", UploadError: "denied"},
			"r2, 0 blocks (0 B) waiting to upload: denied",
		},
	}
	for _, c := range cases {
		if got := formatSpaceStorage(c.storage); got != c.want {
			t.Errorf("formatSpaceStorage(%v) = %q, want %q", c.storage, got, c.want)
		}
	}
}
