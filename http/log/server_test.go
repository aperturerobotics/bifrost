//go:build !js

package httplog

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestHTTPLogServer tests logging http requests to logrus.
func TestHTTPLogServer(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	handler := http.NewServeMux()
	handler.HandleFunc("/test", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(200)
		_, _ = rw.Write([]byte("hello world!\n"))
	})

	srv := httptest.NewServer(LoggingMiddleware(handler, le, LoggingMiddlewareOpts{UserAgent: true}))
	defer srv.Close()
	baseURL, _ := url.Parse(srv.URL)
	baseURL = baseURL.JoinPath("test")

	// Create the client
	client := srv.Client()

	// Get the test endpoint
	resp, err := client.Get(baseURL.String())
	if err != nil {
		t.Fatal(err.Error())
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err.Error())
	}
	if string(data) != "hello world!\n" || resp.StatusCode != 200 {
		t.Fail()
	}
}
