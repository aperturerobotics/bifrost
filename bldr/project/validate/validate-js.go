//go:build js

package bldr_project_validate

import (
	"context"
	"io/fs"

	"github.com/pkg/errors"
	bldr_project "github.com/s4wave/spacewave/bldr/project"
)

// Validate is unavailable in the browser, which cannot evaluate bldr.star.
func Validate(ctx context.Context, fsys fs.FS) (*Validation, error) {
	return nil, errors.New("plugin validation requires the native app")
}

// ValidateProject is unavailable until the browser can evaluate bldr.star.
func ValidateProject(ctx context.Context, fsys fs.FS) (*bldr_project.ProjectConfig, *Validation, error) {
	return nil, nil, errors.New("plugin validation requires the native app")
}
