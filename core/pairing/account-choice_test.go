package pairing

import (
	"net"
	"testing"

	stream_packet "github.com/s4wave/spacewave/net/stream/packet"
)

// TestPairingRejectsChangedAccountChoice checks the authenticated account
// descriptions against the remote proposal before any identity or grant exchange.
func TestPairingRejectsChangedAccountChoice(t *testing.T) {
	for _, changed := range []string{"offered account", "receiving account", "unknown outcome"} {
		t.Run(changed, func(t *testing.T) {
			// Keep the original descriptions observed by the code creator.
			choice := &AccountChoice{
				OfferedAccount:   &AccountOffer{AccountId: "first", OperationId: "first-operation"},
				ReceivingAccount: &AccountOffer{AccountId: "second", OperationId: "second-operation"},
			}
			proposal := choice.CloneVT()
			proposal.Outcome = AccountOutcome_AccountOutcome_SIGN_IN_OFFERED
			switch changed {
			case "offered account":
				proposal.OfferedAccount.AccountId = "substituted"
			case "receiving account":
				proposal.ReceivingAccount.AccountId = "substituted"
			case "unknown outcome":
				proposal.Outcome = AccountOutcome(100)
			}

			// Submit a changed proposal through the actual framed duplex protocol.
			left, right := net.Pipe()
			t.Cleanup(func() { _ = left.Close(); _ = right.Close() })
			sent := make(chan error, 1)
			go func() {
				stream := stream_packet.NewSession(right, 4096)
				sent <- stream.SendMsg(&Frame{Body: &Frame_Choice{Choice: proposal}})
				_ = right.Close()
			}()
			active := &attempt{offering: true}
			engine := &Engine{active: active}
			if err := engine.chooseAccount(t.Context(), active, stream_packet.NewSession(left, 4096), choice); err == nil {
				t.Fatal("changed account choice reached approval")
			}
			if err := <-sent; err != nil {
				t.Fatal(err)
			}
		})
	}
}
