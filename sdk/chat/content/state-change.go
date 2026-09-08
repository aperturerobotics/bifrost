package spacewave_chat_content

import (
	"strings"
	"unicode/utf8"

	"github.com/aperturerobotics/fastjson"
	"github.com/pkg/errors"
)

// Validate checks the state identity and bounded external JSON object.
func (s *ChatStateChange) Validate() error {
	// NUL separates the two external identities in the current-state graph label.
	if s.GetType() == "" || len(s.GetType()) > 255 || len(s.GetStateKey()) > 255 || strings.ContainsRune(s.GetType(), 0) || strings.ContainsRune(s.GetStateKey(), 0) || !utf8.ValidString(s.GetType()) || !utf8.ValidString(s.GetStateKey()) {
		return errors.New("chat state requires a bounded type and state key")
	}

	// State payloads stay public and must be complete objects before storage.
	if len(s.GetContentJson()) > 64*1024 || !utf8.ValidString(s.GetContentJson()) {
		return errors.New("chat state content exceeds its bounds or is not UTF-8")
	}
	value, err := fastjson.Parse(s.GetContentJson())
	if err != nil {
		return errors.Wrap(err, "parse chat state content")
	}
	if value.Type() != fastjson.TypeObject {
		return errors.New("chat state content must be a JSON object")
	}
	return nil
}
