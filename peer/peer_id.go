package peer

import (
	ic "github.com/libp2p/go-libp2p-core/crypto"
	ip "github.com/libp2p/go-libp2p-core/peer"
)

// ID is a peer identifier.
type ID = ip.ID

// IDFromString cast a string to ID type, and validate
// the id to make sure it is a multihash.
func IDFromString(s string) (ID, error) {
	return ip.IDFromString(s)
}

// IDFromBytes cast a string to ID type, and validate
// the id to make sure it is a multihash.
func IDFromBytes(b []byte) (ID, error) {
	return ip.IDFromBytes(b)
}

// IDB58Decode returns a b58-decoded Peer
func IDB58Decode(s string) (ID, error) {
	return ip.IDB58Decode(s)
}

// IDB58Encode returns b58-encoded string
func IDB58Encode(id ID) string {
	return ip.IDB58Encode(id)
}

// IDHexDecode returns a hex-decoded Peer
func IDHexDecode(s string) (ID, error) {
	return ip.IDHexDecode(s)
}

// IDHexEncode returns hex-encoded string
func IDHexEncode(id ID) string {
	return ip.IDHexEncode(id)
}

// IDFromPublicKey returns the Peer ID corresponding to pk
func IDFromPublicKey(pk ic.PubKey) (ID, error) {
	return ip.IDFromPublicKey(pk)
}

// IDFromPrivateKey returns the Peer ID corresponding to sk
func IDFromPrivateKey(sk ic.PrivKey) (ID, error) {
	return ip.IDFromPrivateKey(sk)
}
