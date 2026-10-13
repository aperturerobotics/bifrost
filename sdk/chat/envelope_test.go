package spacewave_chat

import (
	"bytes"
	"testing"

	"github.com/aperturerobotics/protobuf-go-lite/types/known/timestamppb"
	"github.com/s4wave/spacewave/db/world"
	db_world_testbed "github.com/s4wave/spacewave/db/world/testbed"
	chat_rpc "github.com/s4wave/spacewave/sdk/chat/rpc"
)

// TestEnvelopeStoredOnce checks that an envelope is stored once under its ID, whoever
// delivers it, and that its redaction replaces the bytes with the redacted form.
func TestEnvelopeStoredOnce(t *testing.T) {
	// Start a World testbed.
	ctx := t.Context()
	tb, err := db_world_testbed.Default(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tb.Release)

	// Create the channel and its two writers.
	ws := world.NewEngineWorldState(tb.Engine, true)
	const channelKey = "chat/channel/envelope"
	if _, _, err := ws.ApplyWorldOp(ctx, &CreateChatChannelOp{ObjectKey: channelKey, Name: "Envelope", Timestamp: timestamppb.Now()}, tb.Volume.GetPeerID()); err != nil {
		t.Fatal(err)
	}
	alice := newChatResourceForPerson(t, ws, tb.Engine, channelKey, "alice-device", "alice")
	t.Cleanup(alice.Close)
	bob := newChatResourceForPerson(t, ws, tb.Engine, channelKey, "bob-device", "bob")
	t.Cleanup(bob.Close)

	// Alice stores an envelope with the content read from it.
	envelope := &ChatEnvelope{Id: "$event", Data: []byte(`{"type":"m.room.message"}`)}
	request := &chat_rpc.SendMessageRequest{
		Content:  &ChatMessageContent{Content: &ChatMessageContent_Event{Event: &ChatEvent{Type: "m.room.message", ContentJson: `{"body":"hi"}`}}},
		Envelope: envelope,
	}
	sent, err := alice.SendMessage(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	wantKey, err := EnvelopeMessageKey(channelKey, envelope.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if sent.GetMessageKey() != wantKey {
		t.Fatalf("envelope key %q, want %q", sent.GetMessageKey(), wantKey)
	}

	// Bob delivers the same envelope and reaches the accepted message.
	again, err := bob.SendMessage(ctx, request)
	if err != nil || again.GetMessageKey() != sent.GetMessageKey() {
		t.Fatalf("repeated envelope: %v %v", again, err)
	}
	info, err := alice.GetChannelInfo(ctx, &chat_rpc.GetChannelInfoRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if info.GetMessageCount() != 1 {
		t.Fatalf("channel holds %d messages, want 1", info.GetMessageCount())
	}
	read, err := alice.GetMessage(ctx, &chat_rpc.GetMessageRequest{MessageKey: sent.GetMessageKey()})
	if err != nil {
		t.Fatal(err)
	}
	if !read.GetMessage().GetEnvelope().EqualVT(envelope) {
		t.Fatalf("envelope read back as %v", read.GetMessage().GetEnvelope())
	}

	// A redaction without the redacted envelope is rejected.
	redaction := &ChatRedaction{TargetKey: sent.GetMessageKey()}
	redact := &chat_rpc.SendMessageRequest{Content: &ChatMessageContent{Content: &ChatMessageContent_Redaction{Redaction: redaction}}}
	if _, err := alice.SendMessage(ctx, redact); err == nil {
		t.Fatal("redacted an envelope without its redacted form")
	}

	// With it, the stored bytes are replaced.
	redacted := []byte(`{}`)
	redaction.Envelope = &ChatEnvelope{Id: envelope.GetId(), Data: redacted}
	if _, err := alice.SendMessage(ctx, redact); err != nil {
		t.Fatal(err)
	}
	read, err = alice.GetMessage(ctx, &chat_rpc.GetMessageRequest{MessageKey: sent.GetMessageKey()})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(read.GetMessage().GetEnvelope().GetData(), redacted) {
		t.Fatalf("redacted envelope holds %q", read.GetMessage().GetEnvelope().GetData())
	}
}
