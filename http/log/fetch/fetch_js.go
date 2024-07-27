//go:build js

package httplog_fetch

import (
	"net/textproto"
	"slices"
	"time"

	fetch "github.com/aperturerobotics/util/js/fetch"
	"github.com/sirupsen/logrus"
)

// logHeaders is the set of headers to attach to the logger as fields.
var logHeaders = []string{"range", "content-range", "content-type", "content-length", "accept"}

// Fetch uses the JS Fetch API to make requests with logging.
//
// if le is nil, all logging is disabled.
// if verbose is set, both successful and failed calls are logged.
func Fetch(le *logrus.Entry, url string, opts *fetch.Opts, verbose bool) (*fetch.Response, error) {
	// Request details
	if le != nil {
		method := "GET"
		if opts != nil && opts.Method != "" {
			method = opts.Method
		}

		le = le.WithFields(logrus.Fields{
			"method": method,
			"url":    url,
		})

		if opts != nil && opts.Header != nil {
			// Parse and log some headers from the request
			for hdr, hdrVal := range opts.Header {
				hdr = fetch.CanonicalHeaderKey(hdr)
				if slices.Contains(logHeaders, hdr) {
					le = le.WithField(hdr, hdrVal)
				}
			}
		}

		if verbose {
			le.Debug("starting request")
		}
	}

	startTime := time.Now()
	resp, err := fetch.Fetch(url, opts)
	duration := time.Since(startTime)

	if le != nil {
		mapSize := 1
		if resp != nil {
			mapSize += 1 + min(len(resp.Header), len(logHeaders))
		}
		fields := make(logrus.Fields, mapSize)
		fields["dur"] = duration.String()
		if resp != nil {
			fields["status"] = resp.Status
			for hdr, hdrVal := range resp.Header {
				hdr = textproto.CanonicalMIMEHeaderKey(hdr)
				if slices.Contains(logHeaders, hdr) {
					fields[hdr] = hdrVal
				}
			}
		}

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
