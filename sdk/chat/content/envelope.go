package spacewave_chat_content

import (
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// maxEnvelopePrelude bounds the events an envelope carries before its own.
const maxEnvelopePrelude = 16

// maxEnvelopeStateIDs bounds the IDs of an envelope's state.
const maxEnvelopeStateIDs = 16384

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
		if len(event.GetRedactedData()) != 0 || len(event.GetPrelude()) != 0 || event.GetDetached() || len(event.GetStateIds()) != 0 {
			return errors.New("chat envelope prelude event carries only an ID and data")
		}
	}

	// Check the state, which a detached message does not have.
	if e.GetDetached() && (len(e.GetPrelude()) != 0 || len(e.GetStateIds()) != 0) {
		return errors.New("detached chat envelope carries no prelude or state")
	}
	if len(e.GetStateIds()) > maxEnvelopeStateIDs {
		return errors.Errorf("chat envelope state exceeds %d IDs", maxEnvelopeStateIDs)
	}
	for _, id := range e.GetStateIds() {
		if !validID(id) {
			return errors.New("chat envelope state requires bounded protocol IDs")
		}
	}
	return nil
}

// validateEvent checks one event's protocol ID and wire bytes.
func (e *ChatEnvelope) validateEvent() error {
	if !validID(e.GetId()) {
		return errors.New("chat envelope requires a bounded protocol ID")
	}
	if len(e.GetData()) == 0 || len(e.GetData()) > 64*1024 {
		return errors.New("chat envelope data is empty or exceeds 64 KiB")
	}
	return nil
}

// validID reports whether id is a nonempty protocol ID of at most 255 bytes of
// UTF-8 without NUL.
func validID(id string) bool {
	return id != "" && len(id) <= 255 && !strings.ContainsRune(id, 0) && utf8.ValidString(id)
}
