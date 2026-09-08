package spacewave_chat

import (
	"testing"

	"github.com/aperturerobotics/protobuf-go-lite/types/known/timestamppb"
	"github.com/s4wave/spacewave/db/world"
	db_world_testbed "github.com/s4wave/spacewave/db/world/testbed"
	spacewave_chat_rpc "github.com/s4wave/spacewave/sdk/chat/rpc"
)

// TestChatStateHistory checks current state, immutable history, and retry ordering together.
func TestChatStateHistory(t *testing.T) {
	// Create initial state through the production World operation.
	ctx := t.Context()
	wtb, err := db_world_testbed.Default(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(wtb.Release)
	ws := world.NewEngineWorldState(wtb.Engine, true)
	timestamp := timestamppb.Now()
	sender := wtb.Volume.GetPeerID()
	_, _, err = ws.ApplyWorldOp(ctx, &CreateChatChannelOp{
		ObjectKey: GeneralChannelKey,
		Name:      "General",
		Timestamp: timestamp,
		InitialState: []*ChatStateChange{
			{Type: "m.room.create", ContentJson: `{"room_version":"11"}`},
			{Type: "m.room.topic", ContentJson: `{"topic":"First"}`},
		},
	}, sender)
	if err != nil {
		t.Fatal(err)
	}
	resource := NewChatResource(ws, wtb.Engine, GeneralChannelKey, sender.String())
	history, err := resource.ListMessages(ctx, &spacewave_chat_rpc.ListMessagesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(history.GetMessages()) != 2 {
		t.Fatalf("initial history = %d messages, want 2", len(history.GetMessages()))
	}
	for _, message := range history.GetMessages() {
		if !message.GetCreatedAt().EqualVT(timestamp) {
			t.Fatal("initial event timestamp differs from channel creation")
		}
		if message.GetSenderPeerId() != sender.String() {
			t.Fatal("initial event lost its authenticated author")
		}
		if message.GetText() == "" {
			t.Fatal("state event has no readable native history summary")
		}
	}

	// A later state replaces the current value while both accepted messages remain addressable.
	request := &spacewave_chat_rpc.SendMessageRequest{
		TransactionId: "topic-second",
		Content: &ChatMessageContent{Content: &ChatMessageContent_StateChange{
			StateChange: &ChatStateChange{Type: "m.room.topic", ContentJson: `{"topic":"Second"}`},
		}},
	}
	second, err := resource.SendMessage(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	thirdRequest := request.CloneVT()
	thirdRequest.TransactionId = ""
	thirdRequest.Content.GetStateChange().ContentJson = `{"topic":"Third"}`
	third, err := resource.SendMessage(ctx, thirdRequest)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := resource.SendMessage(ctx, thirdRequest)
	if err != nil {
		t.Fatal(err)
	}
	if repeated.GetMessageKey() != third.GetMessageKey() {
		t.Fatal("unchanged state allocated a different event identity")
	}
	retried, err := resource.SendMessage(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if retried.GetMessageKey() != second.GetMessageKey() {
		t.Fatal("old transaction retry lost its accepted event")
	}

	// A fresh attachment sees the latest state with the exact history event identity.
	reader := NewChatResource(ws, nil, GeneralChannelKey, "")
	current, err := reader.GetState(ctx, &spacewave_chat_rpc.GetStateRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(current.GetMessages()) != 2 {
		t.Fatalf("current state = %d messages, want 2", len(current.GetMessages()))
	}
	if current.GetMessages()[1].GetObjectKey() != third.GetMessageKey() {
		t.Fatal("old retry overwrote current state")
	}
	info, err := reader.GetChannelInfo(ctx, &spacewave_chat_rpc.GetChannelInfoRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if info.GetTopic() != "Third" {
		t.Fatalf("native topic = %q, want Third", info.GetTopic())
	}
	if info.GetMessageCount() != 4 {
		t.Fatalf("history count = %d, want 4", info.GetMessageCount())
	}
	old, err := reader.GetMessage(ctx, &spacewave_chat_rpc.GetMessageRequest{MessageKey: history.GetMessages()[1].GetObjectKey()})
	if err != nil {
		t.Fatal(err)
	}
	if old.GetMessage().GetContent().GetStateChange().GetContentJson() != `{"topic":"First"}` {
		t.Fatal("state replacement changed historical event content")
	}

	// Invalid native metadata rolls back the whole append, including its history position.
	thirdRequest.Content.GetStateChange().ContentJson = `{"topic":false}`
	if _, err := resource.SendMessage(ctx, thirdRequest); err == nil {
		t.Fatal("accepted a non-string topic")
	}
	after, err := reader.GetChannelInfo(ctx, &spacewave_chat_rpc.GetChannelInfoRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if !info.EqualVT(after) {
		t.Fatal("rejected state changed channel metadata or history extent")
	}
}

// TestChatStateEncryptionAndCreationRollback checks policy changes and atomic initialization.
func TestChatStateEncryptionAndCreationRollback(t *testing.T) {
	// Reject the entire creation when one initial event violates native metadata policy.
	ctx := t.Context()
	wtb, err := db_world_testbed.Default(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(wtb.Release)
	ws := world.NewEngineWorldState(wtb.Engine, true)
	operation := &CreateChatChannelOp{
		ObjectKey: GeneralChannelKey,
		Timestamp: timestamppb.Now(),
		InitialState: []*ChatStateChange{
			{Type: "m.room.create", ContentJson: `{}`},
			{Type: "m.room.topic", ContentJson: `{"topic":false}`},
		},
	}
	if _, _, err := ws.ApplyWorldOp(ctx, operation, wtb.Volume.GetPeerID()); err == nil {
		t.Fatal("accepted invalid initial metadata")
	}
	object, found, err := ws.GetObject(ctx, GeneralChannelKey)
	world.ReleaseObjectState(object)
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("rejected initialization left a partial channel")
	}

	// Enabling encryption changes the native body policy without hiding public state.
	operation.InitialState = operation.InitialState[:1]
	if _, _, err := ws.ApplyWorldOp(ctx, operation, wtb.Volume.GetPeerID()); err != nil {
		t.Fatal(err)
	}
	resource := NewChatResource(ws, wtb.Engine, GeneralChannelKey, wtb.Volume.GetPeerID().String())
	plaintext := &spacewave_chat_rpc.SendMessageRequest{Text: "accepted before encryption", TransactionId: "plaintext"}
	accepted, err := resource.SendMessage(ctx, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	encryption := &spacewave_chat_rpc.SendMessageRequest{Content: &ChatMessageContent{Content: &ChatMessageContent_StateChange{
		StateChange: &ChatStateChange{Type: "m.room.encryption", ContentJson: `{"algorithm":"m.megolm.v1.aes-sha2"}`},
	}}}
	if _, err := resource.SendMessage(ctx, encryption); err != nil {
		t.Fatal(err)
	}
	retried, err := resource.SendMessage(ctx, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if retried.GetMessageKey() != accepted.GetMessageKey() {
		t.Fatal("encryption activation changed an already accepted send result")
	}
	if _, err := resource.SendMessage(ctx, &spacewave_chat_rpc.SendMessageRequest{Text: "unprotected"}); err == nil {
		t.Fatal("encryption state did not enforce native body policy")
	}
	request := &spacewave_chat_rpc.SendMessageRequest{Content: &ChatMessageContent{Content: &ChatMessageContent_StateChange{
		StateChange: &ChatStateChange{Type: "m.room.topic", ContentJson: `{"topic":"Public metadata"}`},
	}}}
	if _, err := resource.SendMessage(ctx, request); err != nil {
		t.Fatal(err)
	}
	request.Content.GetStateChange().Type = "m.room.encryption"
	request.Content.GetStateChange().ContentJson = `{"algorithm":"other"}`
	if _, err := resource.SendMessage(ctx, request); err == nil {
		t.Fatal("accepted an encryption algorithm replacement")
	}
	request.Content.GetStateChange().Type = "m.room.create"
	request.Content.GetStateChange().ContentJson = `{"room_version":"12"}`
	if _, err := resource.SendMessage(ctx, request); err == nil {
		t.Fatal("accepted replacement creation state")
	}
}
