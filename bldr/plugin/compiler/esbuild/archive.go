package esbuild

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/pkg/errors"
)

// readArchive exposes regular npm package files within the remaining build
// allowance. Archive paths cannot escape the package or introduce symlinks.
func readArchive(ctx context.Context, data []byte, remaining *int64) (map[string][]byte, error) {
	// Open the npm tarball without evaluating any package scripts.
	compressed, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer compressed.Close()

	// Charge archive headers, padding and trailing bytes as well as file contents.
	limited := &io.LimitedReader{R: compressed, N: *remaining + 1}
	archive := tar.NewReader(limited)
	files := make(map[string][]byte)

	// Bound all expanded files before making the package available to resolution.
	for {
		// Stop archive reads when the build is canceled or the archive ends.
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		header, err := archive.Next()
		if errors.Is(err, io.EOF) {
			// Consume the archive tail to enforce the bound and verify the gzip trailer.
			if _, err := io.Copy(io.Discard, limited); err != nil {
				return nil, err
			}
			used := *remaining + 1 - limited.N
			if used > *remaining {
				return nil, errors.New("npm archives exceed expansion bound")
			}
			*remaining -= used
			return files, nil
		}
		if err != nil {
			return nil, err
		}

		// Normalize only paths below npm's package directory.
		name := path.Clean(header.Name)
		if name == "package" && header.Typeflag == tar.TypeDir {
			continue
		}
		name, valid := strings.CutPrefix(name, "package/")
		if !valid || !fs.ValidPath(name) || strings.Contains(name, "\\") {
			return nil, errors.New("npm archive path escapes package")
		}
		if header.Typeflag == tar.TypeDir {
			continue
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			return nil, errors.New("npm archive contains a non-regular file")
		}
		if header.Size > limited.N {
			return nil, errors.New("npm archives exceed expansion bound")
		}

		// Retain complete file bytes while the reader bounds all archive records.
		contents, err := io.ReadAll(archive)
		if err != nil {
			return nil, err
		}
		files[name] = contents
	}
}
