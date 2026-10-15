//go:build !goscript

package plugin_host_wazero_quickjs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	quickjswasi "github.com/aperturerobotics/go-quickjs-wasi-reactor"
	quickjs "github.com/aperturerobotics/go-quickjs-wasi-reactor/wazero-quickjs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// bytecodeExt is the file extension of a cached module.
const bytecodeExt = ".qbc"

// reactorSum identifies the QuickJS reactor build that reads and writes the
// cached bytecode.
var reactorSum = sync.OnceValue(func() [sha256.Size]byte {
	return sha256.Sum256(quickjswasi.QuickJSWASM)
})

// bytecodeCache stores the bytecode of plugin modules compiled by the QuickJS
// reactor, so a plugin start links its modules without parsing their source.
//
// Each plugin has a directory holding one file per module, named by the
// SHA-256 of the reactor build, the module name and the module source.
// QuickJS trusts the bytecode it reads, so only this host writes the cache.
type bytecodeCache struct {
	// dir is the cache root directory.
	dir string
}

// loadModules registers each ES module in the plugin dist with the QuickJS
// instance, from cached bytecode or by compiling and caching it. It then
// removes the plugin's cached modules that the dist no longer contains.
//
// A failure to cache is logged, and a module that does not compile is left
// for the QuickJS module loader to report if the plugin imports it.
func (c *bytecodeCache) loadModules(ctx context.Context, le *logrus.Entry, qjs *quickjs.QuickJS, pluginID string, dist fs.FS) error {
	// Create the plugin's cache directory.
	start := time.Now()
	pluginDir := filepath.Join(c.dir, pluginID)
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		return err
	}

	// Register every module in the dist.
	used := make(map[string]struct{})
	var hits, misses int
	err := fs.WalkDir(dist, ".", func(distPath string, d fs.DirEntry, err error) error {
		// Read the module source and name its cache entry.
		if err != nil || d.IsDir() || !isModulePath(distPath) {
			return err
		}
		src, err := fs.ReadFile(dist, distPath)
		if err != nil {
			return err
		}
		name := path.Join(DistFsMount, distPath)
		file := bytecodeFileName(name, src)
		used[file] = struct{}{}

		// Load the cached bytecode, discarding an entry QuickJS rejects.
		cachePath := filepath.Join(pluginDir, file)
		bytecode, err := os.ReadFile(cachePath)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		if err == nil {
			loadErr := qjs.LoadModule(ctx, bytecode)
			if loadErr == nil {
				hits++
				return nil
			}
			le.WithError(loadErr).WithField("module", name).Warn("discarding cached quickjs module bytecode")
			_ = os.Remove(cachePath)
		}

		// Compile the module and cache its bytecode.
		misses++
		bytecode, err = qjs.CompileModule(ctx, string(src), name)
		if err != nil {
			le.WithError(err).WithField("module", name).Warn("cannot compile quickjs module bytecode")
			return nil
		}
		if err := writeFileAtomic(cachePath, bytecode); err != nil {
			le.WithError(err).WithField("module", name).Warn("cannot cache quickjs module bytecode")
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Remove cached modules this dist does not use.
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		file := entry.Name()
		if _, ok := used[file]; !ok && strings.HasSuffix(file, bytecodeExt) {
			_ = os.Remove(filepath.Join(pluginDir, file))
		}
	}

	// Report the cache use and the time to register the modules.
	le.WithFields(logrus.Fields{
		"hits":     hits,
		"misses":   misses,
		"duration": time.Since(start).String(),
	}).Debug("loaded quickjs module bytecode")
	return nil
}

// deletePlugin removes the cached modules of a plugin.
func (c *bytecodeCache) deletePlugin(pluginID string) error {
	return os.RemoveAll(filepath.Join(c.dir, pluginID))
}

// isModulePath reports whether a dist path names an ES module.
func isModulePath(p string) bool {
	return strings.HasSuffix(p, ".mjs") || strings.HasSuffix(p, ".js")
}

// bytecodeFileName returns the cache file name of a module.
func bytecodeFileName(name string, src []byte) string {
	// Hash the reactor build, the module name and the source.
	sum := reactorSum()
	h := sha256.New()
	_, _ = h.Write(sum[:])
	_, _ = h.Write([]byte(name))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write(src)
	return hex.EncodeToString(h.Sum(nil)) + bytecodeExt
}

// writeFileAtomic writes a file through a temporary file and a rename, so a
// concurrent reader sees the whole file or none of it.
func writeFileAtomic(name string, data []byte) error {
	// Write the data to a temporary file beside the target.
	tmp, err := os.CreateTemp(filepath.Dir(name), filepath.Base(name)+".*.tmp")
	if err != nil {
		return err
	}
	_, err = tmp.Write(data)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}

	// Rename it over the target, or remove it on failure.
	if err == nil {
		err = os.Rename(tmp.Name(), name)
	}
	if err != nil {
		_ = os.Remove(tmp.Name())
	}
	return err
}
