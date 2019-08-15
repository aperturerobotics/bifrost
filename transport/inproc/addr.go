package inproc

import (
	"net"
	"strings"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/pkg/errors"
)

var scheme = "inproc://"

type addr struct {
	peerID peer.ID
	str    string
}

// newAddr builds a new addr
func newAddr(peerID peer.ID) *addr {
	return &addr{
		peerID: peerID,
		str:    scheme + peerID.Pretty(),
	}
}

func parseAddr(addr string) (net.Addr, error) {
	if !strings.HasPrefix(addr, scheme) {
		return nil, errors.Errorf("expected inproc prefix: %s", addr)
	}
	pid, err := peer.IDB58Decode(addr[len(scheme):])
	if err != nil {
		return nil, err
	}
	return newAddr(pid), nil
}

func (a *addr) Network() string {
	return "inproc"
}

func (a *addr) String() string {
	return a.str
}

// _ is a type assertion
var _ net.Addr = ((*addr)(nil))
