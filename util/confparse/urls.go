package confparse

import (
	"net/url"

	"github.com/pkg/errors"
)

// ParseURL parses the url from a string.
// If there is no URL specified, returns nil, nil.
func ParseURL(uri string) (*url.URL, error) {
	if uri == "" {
		return nil, nil
	}
	return url.Parse(uri)
}

// ValidateURL checks if a URL is set and valid.
func ValidateURL(uri string, allowEmpty bool) error {
	url, err := ParseURL(uri)
	if err == nil && (url == nil && !allowEmpty) {
		err = errors.New("url cannot be empty")
	}
	return err
}
