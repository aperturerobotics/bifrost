package spacewave_chat_content

import (
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// maxEnvelopePrelude bounds the events an envelope carries before its own.
const maxEnvelopePrelude = 16

// Validate checks the bounded protocol identities and wire bytes before storage.
func (e *ChatEnvelope) Validate() error {
	// Check the envelope's own event.
	if err := e.validateEvent(); err != nil {
		return err
	}
	if len(e.GetRedactedData()) > 64*1024 {
		return errors.New("chat envelope redacted data exceeds 64 KiB")
	}

	// Check each prelude event, which carries only an ID and data.
	if len(e.GetPrelude()) > maxEnvelopePrelude {
		return errors.Errorf("chat envelope prelude exceeds %d events", maxEnvelopePrelude)
	}
	for _, event := range e.GetPrelude() {
		if err := event.validateEvent(); err != nil {
			return err
		}
		if len(event.GetRedactedData()) != 0 || len(event.GetPrelude()) != 0 {
			return errors.New("chat envelope prelude event carries only an ID and data")
		}
	}
	return nil
}

// validateEvent checks one event's protocol ID and wire bytes.
func (e *ChatEnvelope) validateEvent() error {
	if e.GetId() == "" || len(e.GetId()) > 255 || strings.ContainsRune(e.GetId(), 0) || !utf8.ValidString(e.GetId()) {
		return errors.New("chat envelope requires a bounded protocol ID")
	}
	if len(e.GetData()) == 0 || len(e.GetData()) > 64*1024 {
		return errors.New("chat envelope data is empty or exceeds 64 KiB")
	}
	return nil
}
