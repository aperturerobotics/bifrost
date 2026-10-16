package spacewave_chat

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/world"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
)

// AttachChatEnvelopeOpID identifies a replayable envelope attachment.
const AttachChatEnvelopeOpID = "spacewave-chat/envelope/attach"

// AuthenticatedOperation selects the verified transaction signer on replay.
func (o *AttachChatEnvelopeOp) AuthenticatedOperation() {}

// GetOperationTypeId identifies the attachment for World replay.
func (o *AttachChatEnvelopeOp) GetOperationTypeId() string { return AttachChatEnvelopeOpID }

// Validate requires a message of the channel and a valid envelope.
func (o *AttachChatEnvelopeOp) Validate() error {
	if o.GetObjectKey() == "" {
		return world.ErrEmptyObjectKey
	}
	if suffix, ok := strings.CutPrefix(o.GetMessageKey(), o.GetObjectKey()+"/message/"); !ok || suffix == "" {
		return errors.New("chat envelope attachment names no message of the channel")
	}
	if o.GetEnvelope() == nil {
		return errors.New("chat envelope attachment requires an envelope")
	}
	return o.GetEnvelope().Validate()
}

// ApplyWorldOp attaches the envelope in replay order unless the message holds one.
func (o *AttachChatEnvelopeOp) ApplyWorldOp(ctx context.Context, _ *logrus.Entry, ws world.WorldState, _ peer.ID) (bool, error) {
	// Validate the attachment before reading the message.
	if err := o.Validate(); err != nil {
		return false, err
	}

	// An envelope the message already holds stands.
	message, err := world.LookupObjectBody[ChatMessage](ctx, ws, o.GetMessageKey())
	if err != nil {
		return false, err
	}
	if message.GetEnvelope() != nil {
		return false, nil
	}

	// The envelope's ID must not name a message stored under its envelope key.
	envelopeKey, err := EnvelopeMessageKey(o.GetObjectKey(), o.GetEnvelope().GetId())
	if err != nil {
		return false, err
	}
	object, found, err := ws.GetObject(ctx, envelopeKey)
	world.ReleaseObjectState(object)
	if err != nil {
		return false, err
	}
	if found {
		return false, errors.New("chat envelope ID names another message")
	}

	// Attach the envelope, in its redacted form when the message is redacted,
	// and apply a redaction that named the envelope before it was attached.
	message.Envelope = o.GetEnvelope().CloneVT()
	if message.GetRedactedByKey() != "" {
		message.Envelope.Data, message.Envelope.RedactedData = message.Envelope.GetRedactedData(), nil
	}
	if err := applyPendingRedaction(ctx, ws, o.GetObjectKey(), envelopeKey, message); err != nil {
		return false, err
	}
	return false, writeObjectBody(ctx, ws, o.GetMessageKey(), message)
}

// ApplyWorldObjectOp rejects object-only replay for envelope attachment.
func (o *AttachChatEnvelopeOp) ApplyWorldObjectOp(context.Context, *logrus.Entry, world.ObjectState, peer.ID) (bool, error) {
	return false, world.ErrUnhandledOp
}

// LookupAttachChatEnvelopeOp resolves the replayable attachment operation.
func LookupAttachChatEnvelopeOp(_ context.Context, operationTypeID string) (world.Operation, error) {
	if operationTypeID == AttachChatEnvelopeOpID {
		return &AttachChatEnvelopeOp{}, nil
	}
	return nil, nil
}

var _ world.AuthenticatedOperation = (*AttachChatEnvelopeOp)(nil)
