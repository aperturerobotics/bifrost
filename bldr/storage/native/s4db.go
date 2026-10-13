//go:build !js && !bldr_sqlite

package storage_native

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/bldr/storage"
	volume_controller "github.com/s4wave/spacewave/db/volume/controller"
	volume_s4db "github.com/s4wave/spacewave/db/volume/s4db"
)

// VolumeExt is the file extension of a native Volume.
const VolumeExt = ".s4wave"

// S4db stores each Volume in one s4db file under a root directory.
type S4db struct {
	// verbose logs every store operation.
	verbose bool
	// rootDir holds the Volume files.
	rootDir string
	// cacheBytes bounds the decoded index page cache of each Volume.
	// Zero uses the s4db default.
	cacheBytes int64
	// checkpointMaxBytes bounds the overlay each Volume holds before a
	// checkpoint. Zero uses the s4db default.
	checkpointMaxBytes int64
}

// S4dbOption configures the s4db storage method.
type S4dbOption func(*S4db)

// WithCacheBytes sets the index page cache size of every Volume.
func WithCacheBytes(n int64) S4dbOption {
	return func(s *S4db) { s.cacheBytes = n }
}

// WithCheckpointMaxBytes sets the checkpoint overlay bound of every Volume.
func WithCheckpointMaxBytes(n int64) S4dbOption {
	return func(s *S4db) { s.checkpointMaxBytes = n }
}

// NewS4db constructs the s4db storage method rooted at rootDir.
func NewS4db(verbose bool, rootDir string, opts ...S4dbOption) storage.Storage {
	s := &S4db{verbose: verbose, rootDir: rootDir}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// GetStorageInfo returns StorageInfo.
func (s *S4db) GetStorageInfo() *storage.StorageInfo {
	return &storage.StorageInfo{}
}

// AddFactories adds the factories to the resolver.
func (s *S4db) AddFactories(b bus.Bus, sr *static.Resolver) {
	sr.AddFactory(volume_s4db.NewFactory(b))
}

// VolumePath returns the file of the Volume id under rootDir.
func VolumePath(rootDir, id string) (string, error) {
	// Slashes in the id become underscores.
	filename := strings.ReplaceAll(id, "/", "_") + VolumeExt
	if cleanFilename := filepath.Clean(filename); cleanFilename != filename {
		return "", errors.Errorf("invalid storage id: %s", filename)
	}
	return filepath.Join(rootDir, filename), nil
}

// BuildVolumeConfig creates the volume config for the store ID.
func (s *S4db) BuildVolumeConfig(id string, baseVolCtrlConf *volume_controller.Config) (config.Config, error) {
	path, err := VolumePath(s.rootDir, id)
	if err != nil {
		return nil, err
	}
	return &volume_s4db.Config{
		Path:               path,
		Verbose:            s.verbose,
		VolumeConfig:       baseVolCtrlConf,
		CacheBytes:         s.cacheBytes,
		CheckpointMaxBytes: s.checkpointMaxBytes,
	}, nil
}

// DeleteVolume removes the file of the Volume id.
func (s *S4db) DeleteVolume(id string) error {
	path, err := VolumePath(s.rootDir, id)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func init() {
	storageMethods = append(storageMethods, func(b bus.Bus, rootDir string) []storage.Storage {
		return []storage.Storage{NewS4db(false, rootDir)}
	})
}

// _ is a type assertion
var _ storage.Storage = (*S4db)(nil)
