package httplog

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// DoRequest performs a request with logging.
//
// If verbose=true, logs successful cases as well as errors.
func DoRequest(le *logrus.Entry, client *http.Client, req *http.Request, verbose bool) (*http.Response, error) {
	// Request details
	le = le.WithFields(logrus.Fields{
		"method": req.Method,
		"url":    req.URL.String(),
	})
	if verbose {
		le.Debug("starting request")
	}
	resp, err := client.Do(req)
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
	return resp, err
}
