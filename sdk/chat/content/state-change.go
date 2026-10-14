package spacewave_chat_content

import (
	"strings"
	"unicode/utf8"

	"github.com/aperturerobotics/fastjson"
	"github.com/pkg/errors"
)

// Validate checks the state identity and bounded external JSON object.
func (s *ChatStateChange) Validate() error {
	// Bound the two external identities of the current-state graph label.
	if s.GetType() == "" || !validStateIdentity(s.GetType()) || !validStateIdentity(s.GetStateKey()) {
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

// validStateIdentity reports whether a state type or key is UTF-8 of at most
// 255 characters without NUL, which separates the two in the graph label.
func validStateIdentity(identity string) bool {
	return utf8.ValidString(identity) && utf8.RuneCountInString(identity) <= 255 && !strings.ContainsRune(identity, 0)
}
