//go:build !js

package devtool

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// buildDevtoolBus builds the devtool bus with the command's --var inputs for
// bldr.star to read.
func (a *DevtoolArgs) buildDevtoolBus(ctx context.Context, le *logrus.Entry, repoRoot, stateRoot string, watch bool) (*DevtoolBus, error) {
	vars, err := parseVars(a.Vars.Value())
	if err != nil {
		return nil, err
	}
	return BuildDevtoolBus(ctx, le, repoRoot, stateRoot, watch, vars)
}

// parseVars reads name=value inputs. A repeated name takes its last value.
func parseVars(specs []string) (map[string]string, error) {
	vars := make(map[string]string, len(specs))
	for _, spec := range specs {
		name, value, ok := strings.Cut(spec, "=")
		if !ok || name == "" {
			return nil, errors.Errorf("--var %q must be name=value", spec)
		}
		vars[name] = value
	}
	return vars, nil
}
