package httplog

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// DoRequest performs a request with logging.
//
// If verbose=true, logs successful cases as well as errors.
// le can be nil to disable logging
func DoRequest(le *logrus.Entry, client *http.Client, req *http.Request, verbose bool) (*http.Response, error) {
	return DoRequestWithClient(le, client, req, verbose)
}

// roundTripperClient converts http.RoundTripper to HttpClient.
type roundTripperClient struct {
	rt http.RoundTripper
}

// Do performs the request.
func (r *roundTripperClient) Do(req *http.Request) (*http.Response, error) {
	return r.rt.RoundTrip(req)
}

// _ is a type assertion
var _ HttpClient = (*roundTripperClient)(nil)

// DoRequestWithTransport performs a request with logging.
//
// If verbose=true, logs successful cases as well as errors.
// le can be nil to disable logging
func DoRequestWithTransport(le *logrus.Entry, transport http.RoundTripper, req *http.Request, verbose bool) (*http.Response, error) {
	return DoRequestWithClient(le, &roundTripperClient{rt: transport}, req, verbose)
}

// loggedClient wraps http.Client to HttpClient with a logger.
type loggedClient struct {
	client  HttpClient
	le      *logrus.Entry
	verbose bool
}

// Do performs the request.
func (l *loggedClient) Do(req *http.Request) (*http.Response, error) {
	return DoRequestWithClient(l.le, l.client, req, l.verbose)
}

// _ is a type assertion
var _ HttpClient = (*loggedClient)(nil)

// NewLoggedClient wraps an HttpClient with a logger.
func NewLoggedClient(le *logrus.Entry, client HttpClient, verbose bool) HttpClient {
	return &loggedClient{
		client:  client,
		le:      le,
		verbose: verbose,
	}
}

// HttpClient can perform http requests.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DoRequestWithClient performs a request with logging.
//
// If verbose=true, logs successful cases as well as errors.
// le can be nil to disable logging
func DoRequestWithClient(le *logrus.Entry, client HttpClient, req *http.Request, verbose bool) (*http.Response, error) {
	// Request details
	if le != nil {
		le = le.WithFields(logrus.Fields{
			"method": req.Method,
			"url":    req.URL.String(),
		})

		// Parse and log the Range header from the request
		if rangeHeader := req.Header.Get("Range"); rangeHeader != "" {
			le = le.WithField("range", rangeHeader)
		}

		if verbose {
			le.Debug("starting request")
		}
	}

	var resp *http.Response
	var err error
	startTime := time.Now()
	if client != nil {
		resp, err = client.Do(req)
	} else {
		resp, err = http.DefaultClient.Do(req)
	}
	duration := time.Since(startTime)

	if le != nil {
		fields := make(logrus.Fields, 3)
		fields["dur"] = duration.String()
		if resp != nil {
			fields["status"] = resp.StatusCode

			// Parse and log the Content-Range header from the response
			if contentRangeHeader := resp.Header.Get("Content-Range"); contentRangeHeader != "" {
				fields["response-range"] = contentRangeHeader
			}
		}

		le := le.WithFields(fields)
		if err != nil {
			le.WithError(err).Warn("request errored")
		} else if resp == nil || resp.StatusCode >= 400 {
			le.Warn("request failed")
		} else if verbose {
			le.Debug("request succeeded")
		}
	}

	return resp, err
}

// LoggedRoundTripper is a custom RoundTripper that wraps an existing RoundTripper with a logger.
type LoggedRoundTripper struct {
	transport http.RoundTripper
	le        *logrus.Entry
	verbose   bool
}

// NewLoggedRoundTripper creates a new instance of LoggedRoundTripper.
func NewLoggedRoundTripper(transport http.RoundTripper, le *logrus.Entry, verbose bool) *LoggedRoundTripper {
	return &LoggedRoundTripper{
		transport: transport,
		le:        le,
		verbose:   verbose,
	}
}

// RoundTrip implements the RoundTripper interface.
func (t *LoggedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return DoRequestWithTransport(t.le, t.transport, req, t.verbose)
}
