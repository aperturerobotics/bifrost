package spacewave_chat

import (
	"testing"

	"github.com/aperturerobotics/protobuf-go-lite/types/known/timestamppb"
	"github.com/s4wave/spacewave/db/world"
	db_world_testbed "github.com/s4wave/spacewave/db/world/testbed"
	chat_rpc "github.com/s4wave/spacewave/sdk/chat/rpc"
)

// TestProtocolEventHistory retains extension bodies without changing state or encryption policy.
func TestProtocolEventHistory(t *testing.T) {
	ctx := t.Context()
	tb, err := db_world_testbed.Default(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tb.Release)
	ws := world.NewEngineWorldState(tb.Engine, true)
	const key = "chat/channel/events"
	if _, _, err := ws.ApplyWorldOp(ctx, &CreateChatChannelOp{ObjectKey: key, Name: "Events", Timestamp: timestamppb.Now()}, tb.Volume.GetPeerID()); err != nil {
		t.Fatal(err)
	}
	channel := NewChatResourceForPerson(ws, tb.Engine, key, "device", "person")
	t.Cleanup(channel.Close)
	request := &chat_rpc.SendMessageRequest{TransactionId: "event", Content: &ChatMessageContent{Content: &ChatMessageContent_Event{Event: &ChatEvent{Type: "m.room.test", ContentJson: `{"nested":{"values":[1,true,null,"text"]}}`}}}}
	sent, err := channel.SendMessage(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := channel.SendMessage(ctx, request)
	if err != nil || retry.GetMessageKey() != sent.GetMessageKey() {
		t.Fatalf("retry changed identity: %v %v", retry, err)
	}

	// A new Resource reads the stored body and attribution through ordinary history.
	reader := NewChatResourceForPerson(ws, tb.Engine, key, "other-device", "other-person")
	t.Cleanup(reader.Close)
	history, err := reader.ListMessages(ctx, &chat_rpc.ListMessagesRequest{})
	if err != nil || len(history.GetMessages()) != 1 {
		t.Fatalf("history: %v %v", history, err)
	}
	message := history.GetMessages()[0]
	if !message.GetContent().EqualVT(request.GetContent()) || message.GetPersonPeerId() != "person" || message.GetSenderPeerId() != "device" || message.GetText() != "Unsupported message" {
		t.Fatalf("stored event changed: %v", message)
	}
	state, err := reader.GetState(ctx, &chat_rpc.GetStateRequest{})
	if err != nil || len(state.GetMessages()) != 0 {
		t.Fatalf("timeline event replaced channel state: %v %v", state, err)
	}

	// Conflicting retries and malformed bodies leave history unchanged.
	conflict := request.CloneVT()
	conflict.Content.GetEvent().ContentJson = `{}`
	if _, err := channel.SendMessage(ctx, conflict); err == nil {
		t.Fatal("changed event reused an accepted transaction")
	}
	for _, body := range []string{"[]", "null", "{"} {
		invalid := request.CloneVT()
		invalid.TransactionId = "invalid"
		invalid.Content.GetEvent().ContentJson = body
		if _, err := channel.SendMessage(ctx, invalid); err == nil {
			t.Fatal("accepted malformed event body")
		}
	}
	info, err := reader.GetChannelInfo(ctx, &chat_rpc.GetChannelInfoRequest{})
	if err != nil || info.GetMessageCount() != 1 {
		t.Fatalf("rejected sends advanced history: %v %v", info, err)
	}

	// An extension body cannot bypass encryption by naming another protocol type.
	const encryptedKey = "chat/channel/encrypted-events"
	if _, _, err := ws.ApplyWorldOp(ctx, &CreateChatChannelOp{ObjectKey: encryptedKey, Name: "Encrypted", Timestamp: timestamppb.Now(), EncryptionAlgorithm: "m.megolm.v1.aes-sha2"}, tb.Volume.GetPeerID()); err != nil {
		t.Fatal(err)
	}
	encrypted := NewChatResourceForPerson(ws, tb.Engine, encryptedKey, "device", "person")
	t.Cleanup(encrypted.Close)
	if _, err := encrypted.SendMessage(ctx, request); err == nil {
		t.Fatal("accepted a plaintext extension body in an encrypted channel")
	}
}
