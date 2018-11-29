package transport_controller

import (
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
)

// linkWaiter waits for a link from a specific peer to be opened.
type linkWaiter struct {
	peerID peer.ID
	cb     func(link.Link, bool)
	cbOnce bool
}

// pushLinkWaiter pushes a new waiter for a link with a peer id.
// checks for a link that matches the peer id
// returns nil if callback was called immediately and cbOnce is set
// if cbOnce, only added events will be sent, and only one cb() will be called
// cb added indicates if it is an add or remove event.
// linksMtx should be locked.
func (c *Controller) pushLinkWaiter(
	peerID peer.ID,
	cbOnce bool,
	cb func(lnk link.Link, added bool),
) *linkWaiter {
	peerIDEmpty := peerID == peer.ID("")
	for _, lnk := range c.links {
		if peerIDEmpty || lnk.Link.GetRemotePeer() == peerID {
			go cb(lnk.Link, true)
			if cbOnce {
				return nil
			}
		}
	}

	w := &linkWaiter{peerID: peerID, cb: cb, cbOnce: cbOnce}
	pw := c.linkWaiters[peerID]
	pw = append(pw, w)
	c.linkWaiters[peerID] = pw
	return w
}

// clearLinkWaiter removes waiter for a link
// linksMtx should be locked.
// returns if the waiter was found
func (c *Controller) clearLinkWaiter(w *linkWaiter) bool {
	if w == nil {
		return false
	}

	pid := w.peerID
	pw := c.linkWaiters[pid]
	var found bool
	for i, iw := range pw {
		if iw == w {
			pw[i] = pw[len(pw)-1]
			pw[len(pw)-1] = nil
			pw = pw[:len(pw)-1]
			found = true
			break
		}
	}
	if !found {
		return false
	}

	if len(pw) == 0 {
		delete(c.linkWaiters, pid)
	} else {
		c.linkWaiters[pid] = pw
	}
	return true
}

// resolveLinkWaiters resolves waiters with a link.
// linksMtx should be locked.
func (c *Controller) resolveLinkWaiters(lnk link.Link, added bool) {
	peerID := lnk.GetRemotePeer()
	if pw, ok := c.linkWaiters[peerID]; ok {
		for i := 0; i < len(pw); i++ {
			w := pw[i]
			if w.cbOnce && !added {
				continue
			}
			go w.cb(lnk, added)
			if w.cbOnce {
				pw[i] = pw[len(pw)-1]
				pw[len(pw)-1] = nil
				pw = pw[:len(pw)-1]
				i--
			}
		}
		if len(pw) == 0 {
			delete(c.linkWaiters, peerID)
		} else {
			c.linkWaiters[peerID] = pw
		}
	}
}
