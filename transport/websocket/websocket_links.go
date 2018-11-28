package websocket

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
