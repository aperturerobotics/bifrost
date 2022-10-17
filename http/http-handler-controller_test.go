package bifrost_http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aperturerobotics/bifrost/testbed"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

func TestHTTPHandlerController(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)
	tb, err := testbed.NewTestbed(ctx, le, testbed.TestbedOpts{
		NoEcho: true,
		NoPeer: true,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// start the http server
	busHandler := NewBusHandler(tb.Bus, "test-client")

	// create a http handler
	handler := http.NewServeMux()
	expectedBody := "hello world\n"
	handler.HandleFunc("/foo/bar", func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(200)
		rw.Write([]byte(expectedBody))
	})

	// attach it to the bus
	handlerCtrl := NewHTTPHandlerController(
		controller.NewInfo("bifrost/http/test-handler", semver.MustParse("0.0.1"), "test handler"),
		handler,
		[]string{"/foo"},
		nil,
	)
	relHandler, err := tb.Bus.AddController(ctx, handlerCtrl, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer relHandler()

	// perform a request
	req := httptest.NewRequest("GET", "/foo/bar", nil)
	w := httptest.NewRecorder()
	busHandler.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 status but got %v: %s", resp.StatusCode, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err.Error())
	}
	if string(data) != expectedBody {
		t.Fatalf("expected body %v != %v", string(data), expectedBody)
	}
}
