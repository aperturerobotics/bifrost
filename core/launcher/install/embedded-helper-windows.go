//go:build windows

package launcher_install

import (
	"io/fs"
	"path"
)

// EmbeddedHelper returns the helper executable embedded in this launcher, or
// nil when the build placed none. A launcher without a helper never installs
// itself.
func EmbeddedHelper() []byte {
	data, err := fs.ReadFile(helperFS, path.Join(helperDir, helperName))
	if err != nil {
		return nil
	}
	return data
}
