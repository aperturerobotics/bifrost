package bunlock

import (
	"strings"
	"testing"
)

// TestParseRefusesUnsupportedPins rejects incomplete hashes and non-npm sources.
func TestParseRefusesUnsupportedPins(t *testing.T) {
	// Start from a complete npm tuple with a valid integrity digest.
	valid := string(fixtureLock([]Package{{
		Key:       "answer",
		Name:      "answer",
		Version:   "1.0.0",
		Integrity: fixtureIntegrity([]byte("tarball")),
	}}))
	cases := []struct {
		// name describes the refused lockfile shape.
		name string
		// lock contains the invalid external lockfile.
		lock string
	}{
		{name: "short digest", lock: strings.Replace(valid, fixtureIntegrity([]byte("tarball")), "sha512-short", 1)},
		{name: "other algorithm", lock: strings.Replace(valid, "sha512-", "sha256-", 1)},
		{name: "git source", lock: strings.Replace(valid, "answer@1.0.0", "answer@github:acme/answer", 1)},
		{name: "partial version", lock: strings.Replace(valid, "answer@1.0.0", "answer@1", 1)},
		{name: "unknown version", lock: strings.Replace(valid, `"lockfileVersion":1`, `"lockfileVersion":2`, 1)},
		{name: "missing packages", lock: `{"lockfileVersion":1}`},
		{name: "trusted scripts", lock: strings.Replace(valid, `"packages":`, `"trustedDependencies":[],"packages":`, 1)},
	}

	// Refuse every unsupported form before the fetcher can store bytes.
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse([]byte(tc.lock)); err == nil {
				t.Fatal("accepted an unsupported lock pin")
			}
		})
	}
}

// TestParseExactVersions accepts npm prereleases and dotted build metadata.
func TestParseExactVersions(t *testing.T) {
	// Retain the exact resolved version, including its build metadata.
	for _, version := range []string{"1.2.3", "1.2.3-alpha.1", "1.2.3+build.123", "1.2.3-alpha.1+build.123"} {
		pkg := Package{Key: "answer", Name: "answer", Version: version, Integrity: fixtureIntegrity([]byte("tarball"))}
		packages, err := Parse(fixtureLock([]Package{pkg}))
		if err != nil {
			t.Fatal(err)
		}
		if len(packages) != 1 || packages[0] != pkg {
			t.Fatalf("version %s parsed as %v", version, packages)
		}
	}
}

// TestParsePreservesQuotedContents keeps commas and escaped quotes inside strings.
func TestParsePreservesQuotedContents(t *testing.T) {
	// Add package metadata whose strings contain container-like punctuation.
	lock := string(fixtureLock([]Package{{
		Key:       "answer",
		Name:      "answer",
		Version:   "1.0.0",
		Integrity: fixtureIntegrity([]byte("tarball")),
	}}))
	lock = strings.Replace(lock, "{},", `{"description":"quoted \",] and ,}","dependencies":{"x":"1.0.0",},},`, 1)

	// Decode the real tuple while ignoring unrelated package metadata.
	packages, err := Parse([]byte(lock))
	if err != nil {
		t.Fatal(err)
	}
	if len(packages) != 1 || packages[0].Name != "answer" {
		t.Fatalf("unexpected packages: %v", packages)
	}
}
