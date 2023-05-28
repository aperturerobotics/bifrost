package httplog

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// DoRequest performs a request with verbose logging.
func DoRequest(le *logrus.Entry, client *http.Client, req *http.Request) (*http.Response, error) {
	// Request details
	le = le.WithFields(logrus.Fields{
		"method": req.Method,
		"url":    req.URL.String(),
	})
	resp, err := client.Do(req)
	if resp != nil {
		le = le.WithField("status", resp.StatusCode)
	}
	if err != nil {
		le.WithError(err).Warn("request errored")
	} else if resp.StatusCode >= 400 {
		le.Warn("request failed")
	} else {
		le.Debug("request succeeded")
	}
	return resp, err
}
