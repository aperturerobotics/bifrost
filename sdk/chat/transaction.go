package spacewave_chat

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/pkg/errors"
	spacewave_chat_rpc "github.com/s4wave/spacewave/sdk/chat/rpc"
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

// EnvelopeMessageKey identifies a message another protocol originated by its protocol ID.
// Every sender of the same envelope reaches the same key, so the channel stores it once.
func EnvelopeMessageKey(channelKey, envelopeID string) (string, error) {
	if channelKey == "" || envelopeID == "" {
		return "", errors.New("chat envelope identity requires a channel and envelope ID")
	}
	digest := sha256.Sum256([]byte(envelopeID))
	return channelKey + "/message/env-" + hex.EncodeToString(digest[:]), nil
}

// sendMessageKey returns the key a send request reserves: its envelope's key, or its
// sender-scoped transaction key. A request with neither has no identity yet.
func sendMessageKey(channelKey, senderPeerID string, req *spacewave_chat_rpc.SendMessageRequest) (string, error) {
	if envelope := req.GetEnvelope(); envelope != nil {
		if req.GetTransactionId() != "" {
			return "", errors.New("chat envelope send cannot also carry a transaction ID")
		}
		return EnvelopeMessageKey(channelKey, envelope.GetId())
	}
	if req.GetTransactionId() == "" {
		return "", nil
	}
	return TransactionMessageKey(channelKey, senderPeerID, req.GetTransactionId())
}
