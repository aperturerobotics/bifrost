package cliflags

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	"github.com/aperturerobotics/bifrost/util/confparse"
	"github.com/urfave/cli"
	"strings"
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
