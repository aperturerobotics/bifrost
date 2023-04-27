package confparse

import (
	"testing"

	"github.com/aperturerobotics/timestamp"
)

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *timestamp.Timestamp
		expectError bool
	}{
		{
			name:        "EmptyString",
			input:       "",
			expected:    nil,
			expectError: false,
		},
		{
			name:  "ValidUnixMilliseconds",
			input: "1629048153000",
			expected: &timestamp.Timestamp{
				TimeUnixMs: 1629048153000,
			},
			expectError: false,
		},
		{
			name:  "ValidRFC3339",
			input: "2021-08-15T15:49:13Z",
			expected: &timestamp.Timestamp{
				TimeUnixMs: 1629042553000,
			},
			expectError: false,
		},
		{
			name:        "InvalidInput",
			input:       "invalid",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ParseTimestamp(tt.input)
			if tt.expectError {
				if err == nil {
					t.FailNow()
				}
			} else {
				if err != nil {
					t.Fatal(err.Error())
				}
			}
			if !tt.expected.Equals(actual) {
				t.Fatalf("expected %q got %q", tt.expected, actual)
			}
		})
	}
}
