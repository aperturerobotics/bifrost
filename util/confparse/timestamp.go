package confparse

import (
	"strconv"
	"time"

	"github.com/aperturerobotics/timestamp"
)

// ParseTimestamp parses a timestamp string.
//
// The string can be either a unix time milliseconds or RFC3339 timestamp.
// Returns nil, nil if empty.
func ParseTimestamp(timestampStr string) (*timestamp.Timestamp, error) {
	if timestampStr == "" {
		return nil, nil
	}
	ts := &timestamp.Timestamp{}
	jdat := []byte(strconv.Quote(timestampStr))
	if err := ts.UnmarshalJSON(jdat); err != nil {
		return nil, err
	}
	return ts, nil
}

// MarshalTimestamp marshals a timestamp to a RFC3339 string.
func MarshalTimestamp(ts *timestamp.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.ToTime().Format(time.RFC3339)
}
