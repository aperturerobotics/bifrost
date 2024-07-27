//go:build js && webtests

package httplog_fetch

import (
	"bytes"
	"net/http"
	"testing"

	fetch "github.com/aperturerobotics/util/js/fetch"
	"github.com/sirupsen/logrus"
)

func TestFetch(t *testing.T) {
	// Create a logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Test cases
	testCases := []struct {
		name        string
		url         string
		opts        *fetch.Opts
		verbose     bool
		expectError bool
	}{
		{
			name:    "Successful GET request",
			url:     "https://httpbin.org/get",
			verbose: true,
		},
		{
			name: "POST request with headers",
			url:  "https://httpbin.org/post",
			opts: &fetch.Opts{
				Method: "POST",
				Header: map[string][]string{
					"Content-Type": []string{"application/json"},
				},
				Body: bytes.NewReader([]byte(`{"test": "data"}`)),
			},
			verbose: true,
		},
		{
			name:        "Non-existent URL",
			url:         "https://thisurldoesnotexist.example.com",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			le := logger.WithField("test", tc.name)

			resp, err := Fetch(le, tc.url, tc.opts, tc.verbose)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
				if resp != nil {
					t.Errorf("Expected nil response, but got %v", resp)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if resp == nil {
					t.Fatalf("Expected non-nil response, but got nil")
				}
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status %d, but got %v", http.StatusOK, resp.StatusCode)
				}
			}
		})
	}
}

func TestFetchWithNilLogger(t *testing.T) {
	resp, err := Fetch(nil, "https://httpbin.org/get", nil, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response, but got nil")
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}
