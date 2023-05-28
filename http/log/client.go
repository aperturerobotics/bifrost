package httplog

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// DoRequest performs a request with logging.
//
// If verbose=true, logs successful cases as well as errors.
// le can be nil to disable logging
func DoRequest(le *logrus.Entry, client *http.Client, req *http.Request, verbose bool) (*http.Response, error) {
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
	resp, err := client.Do(req)
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
