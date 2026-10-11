// Package bunlock reads integrity-pinned npm dependencies from bun.lock and
// stores verified tarballs in a caller's block-backed dependency tree.
package bunlock

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"path"
	"regexp"
	"strings"

	"github.com/aperturerobotics/fastjson"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
)

// npmName matches unscoped and scoped npm names without URL metacharacters.
var npmName = regexp.MustCompile(`^(@[A-Za-z0-9._~-]+/)?[A-Za-z0-9._~-]+$`)

// Package is an npm tarball pinned by a bun.lock package entry.
type Package struct {
	// Key is the install path in bun.lock, including nested dependencies.
	Key string
	// Name is the resolved npm package name, which may differ from an alias key.
	Name string
	// Version is the exact resolved npm version.
	Version string
	// URL is the explicit tarball URL, or empty for the caller's default registry.
	URL string
	// Integrity is the canonical SHA-512 SRI string and dependency-tree key.
	Integrity string
}

// Parse reads Bun's version-one lockfile, including its trailing commas.
// Only npm entries with a complete SHA-512 integrity digest are accepted.
// Trusted dependencies are refused because they permit install scripts.
func Parse(data []byte) ([]Package, error) {
	// Require the supported lockfile shape before visiting its packages.
	var parser fastjson.Parser
	lock, err := parser.ParseBytes(withoutTrailingCommas(data))
	if err != nil {
		return nil, errors.Wrap(err, "bun.lock does not parse")
	}
	if lock.GetInt("lockfileVersion") != 1 {
		return nil, errors.New("bun.lock requires lockfileVersion 1")
	}
	if lock.Exists("trustedDependencies") {
		return nil, ErrTrustedDependencies
	}
	packages := lock.GetObject("packages")
	if packages == nil {
		return nil, errors.New("bun.lock requires a packages object")
	}

	// Decode every entry before a fetch can change the dependency tree.
	result := make([]Package, 0, packages.Len())
	packages.Visit(func(key []byte, value *fastjson.Value) {
		if err != nil {
			return
		}
		var pkg Package
		pkg, err = parsePackage(string(key), value.GetArray())
		if err == nil {
			result = append(result, pkg)
		}
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// parsePackage decodes [name@version, tarball URL, package info, integrity].
func parsePackage(key string, fields []*fastjson.Value) (Package, error) {
	// Require the npm tuple and an exact version, excluding file and git entries.
	if len(fields) != 4 {
		ident := key
		if len(fields) != 0 && fields[0].Type() == fastjson.TypeString {
			ident = string(fields[0].GetStringBytes())
		}
		return Package{}, errors.Errorf("bun.lock does not pin %s by its content hash", ident)
	}
	ident := string(fields[0].GetStringBytes())
	at := strings.LastIndexByte(ident, '@')
	if at <= 0 {
		return Package{}, errors.Errorf("bun.lock has an entry for %s without a version", key)
	}
	name, version := ident[:at], ident[at+1:]

	// Require all three numeric version components, retaining npm build metadata.
	core, _, _ := strings.Cut(version, "-")
	core, _, _ = strings.Cut(core, "+")
	if !semver.IsValid("v"+version) || strings.Count(core, ".") != 2 {
		return Package{}, errors.Errorf("bun.lock has an entry for %s without an exact npm version", key)
	}
	if !npmName.MatchString(name) || path.Base(name) == "." || path.Base(name) == ".." {
		return Package{}, errors.Errorf("bun.lock has an invalid npm name for %s", key)
	}
	if fields[1].Type() != fastjson.TypeString || fields[2].Type() != fastjson.TypeObject {
		return Package{}, errors.Errorf("bun.lock has an invalid npm entry for %s", key)
	}

	// Require the full digest before treating its string as a storage key.
	integrity := string(fields[3].GetStringBytes())
	digest, err := base64.StdEncoding.Strict().DecodeString(strings.TrimPrefix(integrity, "sha512-"))
	if !strings.HasPrefix(integrity, "sha512-") || err != nil || len(digest) != sha512.Size {
		return Package{}, errors.Errorf("bun.lock does not pin %s by its content hash", ident)
	}
	return Package{
		Key:       key,
		Name:      name,
		Version:   version,
		URL:       string(fields[1].GetStringBytes()),
		Integrity: "sha512-" + base64.StdEncoding.EncodeToString(digest),
	}, nil
}

// withoutTrailingCommas removes commas before closing brackets and braces,
// preserving string contents.
func withoutTrailingCommas(data []byte) []byte {
	// Track quoted strings so their commas and escapes stay unchanged.
	out := make([]byte, 0, len(data))
	inString, escaped := false, false
	for i, c := range data {
		// Preserve quoted bytes while tracking the closing quote.
		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			out = append(out, c)
			continue
		}

		// Drop only commas whose next token closes a container.
		inString = c == '"'
		if c == ',' {
			rest := bytes.TrimLeft(data[i+1:], " \t\r\n")
			if len(rest) != 0 && (rest[0] == ']' || rest[0] == '}') {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}
