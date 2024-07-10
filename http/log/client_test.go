//go:build !js

package httplog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestDoRequest(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	entry := logrus.NewEntry(logger)

	// Create a test server that responds with 200 OK
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer ts.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", ts.URL, nil)
	resp, err := DoRequest(entry, client, req, true)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %v", resp.StatusCode)
	}

	// Create a test server that responds with 500 Internal Server Error
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer ts2.Close()

	req2, _ := http.NewRequest("GET", ts2.URL, nil)
	resp2, err2 := DoRequest(entry, client, req2, true)

	if err2 != nil {
		t.Errorf("Expected no error, got %v", err2)
	}
	if resp2.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %v", resp2.StatusCode)
	}
}
