package spacewave_chat_content

import (
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// Validate checks the bounded protocol identity and wire bytes before storage.
func (e *ChatEnvelope) Validate() error {
	if e.GetId() == "" || len(e.GetId()) > 255 || strings.ContainsRune(e.GetId(), 0) || !utf8.ValidString(e.GetId()) {
		return errors.New("chat envelope requires a bounded protocol ID")
	}
	if len(e.GetData()) == 0 || len(e.GetData()) > 64*1024 {
		return errors.New("chat envelope data is empty or exceeds 64 KiB")
	}
	return nil
}
