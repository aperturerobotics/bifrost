package spacewave_chat

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/pkg/errors"
)

// TransactionMessageKey identifies a retryable send within its channel World.
// It neither creates a message nor grants access. Callers supply the authenticated sender.
func TransactionMessageKey(channelKey, senderPeerID, transactionID string) (string, error) {
	if channelKey == "" || senderPeerID == "" || transactionID == "" {
		return "", errors.New("chat transaction identity requires a channel, sender, and transaction ID")
	}
	digest := sha256.Sum256([]byte(strconv.Itoa(len(senderPeerID)) + ":" + senderPeerID + transactionID))
	return channelKey + "/message/tx-" + hex.EncodeToString(digest[:]), nil
}
