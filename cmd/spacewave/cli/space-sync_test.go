//go:build !js

package spacewave_cli

import (
	"bytes"
	"testing"
	"time"

	timestamppb "github.com/aperturerobotics/protobuf-go-lite/types/known/timestamppb"
	"github.com/s4wave/spacewave/core/sobject"
)

// TestWriteSyncPeers checks the Sync line and peer table of space info, which
// show a stream that ended and why.
func TestWriteSyncPeers(t *testing.T) {
	// No peers prints only the summary.
	if got := formatSyncPeerCount(nil); got != "no peers since start" {
		t.Fatalf("empty summary = %q", got)
	}
	var empty bytes.Buffer
	writeSyncPeers(&empty, nil)
	if empty.Len() != 0 {
		t.Fatalf("empty table = %q", empty.String())
	}

	// One connected peer and one whose stream ended with an error.
	up, down := time.Unix(1760750000, 0), time.Unix(1760700000, 0)
	peers := []*sobject.SOSyncPeer{
		{PeerId: "peer-a", Connected: true, Since: timestamppb.New(up)},
		{PeerId: "peer-b", Since: timestamppb.New(down), Error: "stream reset"},
	}
	if got := formatSyncPeerCount(peers); got != "1 of 2 peers connected" {
		t.Fatalf("summary = %q", got)
	}
	var out bytes.Buffer
	writeSyncPeers(&out, peers)
	const layout = "2006-01-02 15:04:05"
	want := "\nSync peers\n" +
		"  PEER    STATE         SINCE                ERROR\n" +
		"  peer-a  connected     " + up.Format(layout) + "\n" +
		"  peer-b  disconnected  " + down.Format(layout) + "  stream reset\n"
	if out.String() != want {
		t.Fatalf("table =\n%s\nwant\n%s", out.String(), want)
	}
}
