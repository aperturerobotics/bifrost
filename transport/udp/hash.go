package udp

import (
	"crypto/sha1"
	"encoding/hex"
)

func hashData(data []byte) string {
	d := sha1.Sum(data)
	return hex.EncodeToString(d[:])
}
