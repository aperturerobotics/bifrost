package cli

import (
	"strings"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	"github.com/aperturerobotics/bifrost/util/confparse"
	"github.com/urfave/cli"
)

// parseDialerAddrs parses a dialer map from a string slice
func parseDialerAddrs(ss cli.StringSlice) (map[string]*dialer.DialerOpts, error) {
	m := make(map[string]*dialer.DialerOpts)
	for _, s := range ss {
		pair := strings.Split(s, "@")
		if len(pair) < 2 {
			continue
		}
		pid, err := confparse.ParsePeerID(strings.TrimSpace(pair[0]))
		if err != nil {
			return nil, err
		}
		if pid == peer.ID("") {
			continue
		}
		m[pid.Pretty()] = &dialer.DialerOpts{
			Address: strings.TrimSpace(pair[1]),
		}
	}
	return m, nil
}

// parsePeerIDs parses a peer id list
func parsePeerIDs(ss cli.StringSlice) ([]peer.ID, error) {
	m := make(map[peer.ID]struct{})
	o := make([]peer.ID, 0, len(ss))
	for _, s := range ss {
		pid, err := confparse.ParsePeerID(strings.TrimSpace(s))
		if err != nil {
			return nil, err
		}
		if pid == peer.ID("") {
			continue
		}
		if _, ok := m[pid]; ok {
			continue
		}
		m[pid] = struct{}{}
		o = append(o, pid)
	}
	return o, nil
}
