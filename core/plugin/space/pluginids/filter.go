package pluginids

import (
	"github.com/s4wave/spacewave/bldr/plugin/pluginid"
	"github.com/sirupsen/logrus"
)

// FilterValid drops invalid SpaceSettings plugin IDs.
func FilterValid(le *logrus.Entry, ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	filtered := make([]string, 0, len(ids))
	for _, pid := range ids {
		if err := pluginid.Validate(pid, false); err != nil {
			le.WithError(err).WithField("plugin-id", pid).Warn("ignoring invalid SpaceSettings plugin id")
			continue
		}
		filtered = append(filtered, pid)
	}
	return filtered
}

// Missing returns the IDs in ids that have does not contain, in order.
func Missing(have, ids []string) []string {
	// Index the IDs already present.
	present := make(map[string]struct{}, len(have))
	for _, pid := range have {
		present[pid] = struct{}{}
	}

	// Keep each ID not yet present.
	var missing []string
	for _, pid := range ids {
		if _, ok := present[pid]; !ok {
			missing = append(missing, pid)
		}
	}
	return missing
}
