//go:build !js

package bldr_project_validate

import (
	"io/fs"

	"github.com/aperturerobotics/fastjson"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/bldr/project/bunlock"
	bldr_project_starlark "github.com/s4wave/spacewave/bldr/project/starlark"
)

// lifecycleScripts are the package.json scripts a package manager runs during
// install.
var lifecycleScripts = []string{"preinstall", "install", "postinstall", "preprepare", "prepare", "postprepare"}

// dependencyFields are the package.json fields that declare direct dependencies.
var dependencyFields = []string{"dependencies", "devDependencies", "optionalDependencies", "peerDependencies"}

// checkPackage reads package.json and bun.lock, refuses install scripts and
// dependencies not pinned by hash, and records the locked dependencies.
func (v *Validation) checkPackage(fsys fs.FS) error {
	// A project without package.json installs nothing.
	data, err := bldr_project_starlark.ReadFile(fsys, "package.json")
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if errors.Is(err, bldr_project_starlark.ErrFileTooLarge) {
		v.refuse(RefusalKind_REFUSAL_KIND_CONFIG, "package.json is too large to check.")
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "read package.json")
	}

	// Parse the package manifest.
	var parser fastjson.Parser
	pkg, err := parser.ParseBytes(data)
	if err != nil {
		v.refuse(RefusalKind_REFUSAL_KIND_CONFIG, "package.json does not parse: "+err.Error())
		return nil
	}

	// Refuse scripts and trusted dependencies that run during install.
	for _, script := range lifecycleScripts {
		if pkg.Exists("scripts", script) {
			v.refuse(RefusalKind_REFUSAL_KIND_LIFECYCLE_SCRIPT, "package.json has a "+script+" script, which would run during install.")
		}
	}
	if pkg.Exists("trustedDependencies") {
		v.refuse(RefusalKind_REFUSAL_KIND_LIFECYCLE_SCRIPT, "package.json trusts dependencies to run their install scripts.")
	}

	// Collect the direct dependencies.
	direct := make(map[string]bool)
	for _, field := range dependencyFields {
		if deps := pkg.GetObject(field); deps != nil {
			deps.Visit(func(name []byte, _ *fastjson.Value) {
				direct[string(name)] = true
			})
		}
	}

	// Require bun.lock whenever there is anything to install.
	lock, err := bldr_project_starlark.ReadFile(fsys, "bun.lock")
	if errors.Is(err, fs.ErrNotExist) {
		if len(direct) != 0 {
			v.refuse(RefusalKind_REFUSAL_KIND_UNPINNED_DEPENDENCY, "package.json has dependencies but no bun.lock pins them.")
		}
		return nil
	}
	if errors.Is(err, bldr_project_starlark.ErrFileTooLarge) {
		v.refuse(RefusalKind_REFUSAL_KIND_UNPINNED_DEPENDENCY, "bun.lock is too large to check.")
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "read bun.lock")
	}
	v.checkLock(lock, direct)
	return nil
}

// checkLock records each package bun.lock pins and refuses any it does not
// pin by an integrity hash, such as git, file and workspace packages.
func (v *Validation) checkLock(data []byte, direct map[string]bool) {
	// Use the fetcher's parser so confirmation and downloading accept the same pins.
	packages, err := bunlock.Parse(data)
	if errors.Is(err, bunlock.ErrTrustedDependencies) {
		v.refuse(RefusalKind_REFUSAL_KIND_LIFECYCLE_SCRIPT, err.Error()+".")
		return
	}
	if err != nil {
		v.refuse(RefusalKind_REFUSAL_KIND_UNPINNED_DEPENDENCY, err.Error()+".")
		return
	}

	// Record resolved packages and account for each declared direct dependency.
	pinned := make(map[string]bool, len(packages))
	for _, pkg := range packages {
		pinned[pkg.Key] = true
		v.Dependencies = append(v.Dependencies, &Dependency{
			Name:    pkg.Name,
			Version: pkg.Version,
			Direct:  direct[pkg.Key],
		})
	}

	// Refuse a manifest dependency absent from the lock's top-level package keys.
	for name := range direct {
		if !pinned[name] {
			v.refuse(RefusalKind_REFUSAL_KIND_UNPINNED_DEPENDENCY, "bun.lock does not pin the direct dependency "+name+".")
		}
	}
}
