package kernel

// epollEntry is a file on an epoll set's interest list, added under the
// descriptor fd.
type epollEntry struct {
	// fd is the descriptor the file was added under.
	fd uint32
	// of is the file.
	of *OpenFile
	// events is the events the entry watches, with its EPOLLET and
	// EPOLLONESHOT flags.
	events uint32
	// data is returned with each event.
	data uint64
	// disabled is set once a one-shot entry reports, until a modify.
	disabled bool
	// seen holds the read and write change channels of an edge-triggered
	// entry's last report. Its readiness has not changed while they are
	// open.
	seen [2]<-chan struct{}
}

// poll returns the events that have happened on the file, and the channels
// that close at its next change. An edge-triggered entry reports only after
// its readiness changed since its last report, and a one-shot entry reports
// once.
func (e *epollEntry) poll() (uint32, []<-chan struct{}) {
	// Check the read side, whose hangup is reported unasked, then the write
	// side when watched.
	if e.disabled {
		return 0, nil
	}
	var revents uint32
	var changes []<-chan struct{}
	var current [2]<-chan struct{}
	for i, write := range []bool{false, true} {
		mask := uint32(epollIn)
		if write {
			mask = epollOut
			if e.events&epollOut == 0 {
				continue
			}
		}
		ready, hangup, change := e.of.readiness(write)
		if ready && e.events&mask != 0 {
			revents |= mask
		}
		if hangup {
			revents |= epollHup | e.events&epollRdhup
		}
		if change != nil {
			current[i] = change
			changes = append(changes, change)
		}
	}

	// Hold back an edge-triggered report until the readiness changes.
	if revents == 0 || e.events&epollET != 0 && current == e.seen {
		return 0, changes
	}
	e.seen = current
	e.disabled = e.events&epollOneshot != 0
	return revents, changes
}
