package sobject_sync

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/s4wave/spacewave/core/sobject"
	stream_packet "github.com/s4wave/spacewave/net/stream/packet"
)

// TestCatchupBudgetsRejectWithoutMutation sends excessive history over an
// authenticated stream and checks that neither a prefix nor its head is adopted.
func TestCatchupBudgetsRejectWithoutMutation(t *testing.T) {
	for _, test := range []struct {
		// name identifies the independently enforced protocol budget.
		name string
		// count is the number of untrusted linked entries sent.
		count int
		// padding enlarges each entry without changing cursor continuity.
		padding int
		// perPage chooses a page size, including an intentionally excessive one.
		perPage int
		// oversizedFrame sends only an excessive length prefix, with no payload allocation.
		oversizedFrame bool
	}{
		{name: "page entries", count: maxHistoryPageEntries + 1, perPage: maxHistoryPageEntries + 1},
		{name: "page bytes", count: 1, padding: maxHistoryPageBytes, perPage: 1},
		{name: "suffix entries", count: sobject.MaxConfigSuffixEntries + 1, perPage: maxHistoryPageEntries},
		{name: "suffix bytes", count: 33, padding: 256 * 1024, perPage: 3},
		{name: "snapshot frame", oversizedFrame: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			const soID = "catchup-budget"
			owner, reader := mustKeyPair(t), mustKeyPair(t)
			initial := authenticationState(t, soID, owner, reader)
			local := newAuthenticationPeer(t, soID, owner, initial)
			remote := newAuthenticationPeer(t, soID, reader, initial)
			left, right := net.Pipe()
			t.Cleanup(func() { left.Close(); right.Close() })
			done := make(chan error, 1)
			go func() { done <- local.runStream(ctx, gateLogger(), left, "transport-a", "transport-b") }()
			if _, err := remote.authenticate(ctx, stream_packet.NewSession(right, 64*1024), "transport-b", "transport-a"); err != nil {
				t.Fatal(err)
			}
			session := stream_packet.NewSession(right, maxMessageSize)
			head := &SOSyncMessage{}
			if err := session.RecvMsg(head); err != nil {
				t.Fatal(err)
			}
			if err := session.SendMsg(syncAcknowledgment(head.GetHead().GetRevision())); err != nil {
				t.Fatal(err)
			}

			// Bytes beyond the frame cap must be rejected before a body is read.
			if test.oversizedFrame {
				var prefix [4]byte
				binary.LittleEndian.PutUint32(prefix[:], maxMessageSize+1)
				if _, err := right.Write(prefix[:]); err != nil {
					t.Fatal(err)
				}
			} else {
				// Entries are linked but untrusted; budget rejection precedes signature work.
				cursor := initial.Config.ConfigChainHash
				changes := make([]*sobject.SOConfigChange, 0, test.count)
				for i := range test.count {
					change := &sobject.SOConfigChange{
						ConfigSeqno: uint64(i + 1), PreviousHash: cursor,
						Config: &sobject.SharedObjectConfig{ConfigChainHash: bytes.Repeat([]byte{1}, test.padding)},
					}
					var err error
					cursor, err = sobject.HashSOConfigChange(change)
					if err != nil {
						t.Fatal(err)
					}
					changes = append(changes, change)
				}
				if err := session.SendMsg(&SOSyncMessage{Body: &SOSyncMessage_Head{Head: &SOSyncHead{
					Revision: 1, ConfigHash: cursor, ConfigSeqno: uint64(test.count), RootSeqno: 1, StateHash: bytes.Repeat([]byte{1}, 32),
				}}}); err != nil {
					t.Fatal(err)
				}
				request := &SOSyncMessage{}
				if err := session.RecvMsg(request); err != nil {
					t.Fatal(err)
				}

				// The last page crosses exactly the selected budget and must terminate catch-up.
				cursor = initial.Config.ConfigChainHash
				for len(changes) != 0 {
					count := min(test.perPage, len(changes))
					page := &SOSyncHistoryPage{Revision: 1, Cursor: cursor, Changes: changes[:count]}
					if err := session.SendMsg(&SOSyncMessage{Body: &SOSyncMessage_HistoryPage{HistoryPage: page}}); err != nil {
						t.Fatal(err)
					}
					var err error
					cursor, err = sobject.HashSOConfigChange(changes[count-1])
					if err != nil {
						t.Fatal(err)
					}
					changes = changes[count:]
				}
				response := &SOSyncMessage{}
				if err := session.RecvMsg(response); err != nil || response.GetRecoveryRequired() == nil {
					t.Fatalf("budget recovery response = %v: %v", response, err)
				}
			}
			select {
			case err := <-done:
				if !errors.Is(err, sobject.ErrConfigHistoryUnavailable) {
					t.Fatalf("budget result = %v", err)
				}
			case <-ctx.Done():
				t.Fatal("budget rejection did not stop the stream")
			}
			got, err := local.soHost.GetHostState(ctx)
			if err != nil || !got.EqualVT(initial) {
				t.Fatalf("budget failure changed held state: %v", err)
			}
		})
	}
}
