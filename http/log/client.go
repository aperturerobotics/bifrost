package httplog

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// LoggedRoundTripper is a custom RoundTripper that wraps an existing RoundTripper with a logger.
type LoggedRoundTripper struct {
	transport http.RoundTripper
	le        *logrus.Entry
}

// NewLoggedRoundTripper creates a new instance of LoggedRoundTripper.
func NewLoggedRoundTripper(transport http.RoundTripper, le *logrus.Entry) *LoggedRoundTripper {
	return &LoggedRoundTripper{
		transport: transport,
		le:        le,
	}
}

// RoundTrip implements the RoundTripper interface.
func (t *LoggedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return DoRequestWithTransport(t.le, t.transport, req, false)
}

// DoRequest performs a request with logging.
//
// If verbose=true, logs successful cases as well as errors.
// le can be nil to disable logging
func DoRequest(le *logrus.Entry, client *http.Client, req *http.Request, verbose bool) (*http.Response, error) {
	return DoRequestWithTransport(le, client.Transport, req, verbose)
}

// DoRequestWithTransport performs a request with logging.
//
// If verbose=true, logs successful cases as well as errors.
// le can be nil to disable logging
func DoRequestWithTransport(le *logrus.Entry, transport http.RoundTripper, req *http.Request, verbose bool) (*http.Response, error) {
	// Request details
	if le != nil {
		le = le.WithFields(logrus.Fields{
			"method": req.Method,
			"url":    req.URL.String(),
		})
		if verbose {
			le.Debug("starting request")
		}
	}

	var resp *http.Response
	var err error
	if transport != nil {
		resp, err = transport.RoundTrip(req)
	} else {
		resp, err = http.DefaultTransport.RoundTrip(req)
	}

	if le != nil {
		if resp != nil {
			le = le.WithField("status", resp.StatusCode)
		}
		if err != nil {
			le.WithError(err).Warn("request errored")
		} else if resp.StatusCode >= 400 {
			le.Warn("request failed")
		} else if verbose {
			le.Debug("request succeeded")
		}
	}

	return resp, err
}

// ClientWithLogger wraps an http.Client with a logger.
func ClientWithLogger(client http.Client, le *logrus.Entry) http.Client {
	return http.Client{
		Transport:     NewLoggedRoundTripper(client.Transport, le),
		CheckRedirect: client.CheckRedirect,
		Jar:           client.Jar,
		Timeout:       client.Timeout,
	}
}
