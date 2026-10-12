//go:build !js

package control

import (
	"errors"
	"os"
)

// NoListener reports whether err from dialing a Unix socket proves that no
// listener holds the path: the path is absent or the connection was refused.
// Other errors, such as a permission failure, leave the path's owner unknown.
func NoListener(err error) bool {
	return errors.Is(err, os.ErrNotExist) || errors.Is(err, errRefused)
}
