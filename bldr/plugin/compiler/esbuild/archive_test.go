package esbuild

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"testing"
)

// TestArchiveRefusesPathsAndLinks protects the virtual package filesystem from
// archive records that escape its directory or redirect module reads.
func TestArchiveRefusesPathsAndLinks(t *testing.T) {
	// Exercise npm archive records through the same decoder used by compilation.
	for _, header := range []tar.Header{
		{Name: "package/../../escape.js", Typeflag: tar.TypeReg},
		{Name: "/package/index.js", Typeflag: tar.TypeReg},
		{Name: "package/link.js", Typeflag: tar.TypeSymlink, Linkname: "../../escape.js"},
		{Name: "package/link.js", Typeflag: tar.TypeLink, Linkname: "package/index.js"},
	} {
		t.Run(header.Name+string(header.Typeflag), func(t *testing.T) {
			// Encode the untrusted record as a complete compressed npm archive.
			var output bytes.Buffer
			compressed := gzip.NewWriter(&output)
			archive := tar.NewWriter(compressed)
			if err := archive.WriteHeader(&header); err != nil {
				t.Fatal(err)
			}
			if err := archive.Close(); err != nil {
				t.Fatal(err)
			}
			if err := compressed.Close(); err != nil {
				t.Fatal(err)
			}

			// Refuse it before any redirected or escaped file can reach esbuild.
			remaining := int64(4096)
			if _, err := readArchive(t.Context(), output.Bytes(), &remaining); err == nil {
				t.Fatal("accepted an archive path escape or link")
			}
		})
	}
}

// TestArchiveBoundsTotalExpansion counts empty headers and padding as well as
// source bytes, so many small packages cannot bypass the build's memory bound.
func TestArchiveBoundsTotalExpansion(t *testing.T) {
	// Decode one complete package and retain its remaining build allowance.
	data := npmArchive(t, map[string]string{"index.js": "export default 42"})
	remaining := int64(4095)
	if _, err := readArchive(t.Context(), data, &remaining); err != nil {
		t.Fatal(err)
	}

	// Refuse another package whose total expanded archive crosses the same bound.
	if _, err := readArchive(t.Context(), data, &remaining); err == nil {
		t.Fatal("multiple packages bypassed the total archive expansion bound")
	}
}
