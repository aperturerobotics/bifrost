package websocket

import (
	"github.com/aperturerobotics/bifrost/link"
)

// GetLinks returns the links currently active.
func (u *Transport) GetLinks() (lnks []link.Link) {
	u.linksMtx.Lock()
	defer u.linksMtx.Unlock()

	lnks = make([]link.Link, 0, len(u.links))
	for _, lnk := range u.links {
		lnks = append(lnks, lnk)
	}

	return
}

// handleLinkLost is called when a link is lost.
func (u *Transport) handleLinkLost(addr string, lnk *Link) {
	u.linksMtx.Lock()
	existing := u.links[addr]
	rel := existing == lnk
	if rel {
		delete(u.links, addr)
	}
	u.linksMtx.Unlock()

	if u.handler != nil && rel {
		u.handler.HandleLinkLost(lnk)
	}
}
