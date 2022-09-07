package peer

import (
	ic "github.com/libp2p/go-libp2p/core/crypto"
	ip "github.com/libp2p/go-libp2p/core/peer"
	b58 "github.com/mr-tron/base58/base58"
)

// ID is a peer identifier.
type ID = ip.ID

// IDFromBytes cast a string to ID type, and validate
// the id to make sure it is a multihash.
func IDFromBytes(b []byte) (ID, error) {
	return ip.IDFromBytes(b)
}

// IDB58Decode returns a b58-decoded Peer
func IDB58Decode(s string) (ID, error) {
	return ip.Decode(s)
}

// IDB58Encode returns b58-encoded string
func IDB58Encode(id ID) string {
	return b58.Encode([]byte(id))
}

// IDFromPublicKey returns the Peer ID corresponding to pk
func IDFromPublicKey(pk ic.PubKey) (ID, error) {
	return ip.IDFromPublicKey(pk)
}

// IDFromPrivateKey returns the Peer ID corresponding to sk
func IDFromPrivateKey(sk ic.PrivKey) (ID, error) {
	return ip.IDFromPrivateKey(sk)
}
