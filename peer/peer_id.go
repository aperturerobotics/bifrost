// Package peer defines the peer identifier format
// Based on the libp2p peer format
package peer

import (
	"encoding/hex"
	"strings"

	ic "github.com/libp2p/go-libp2p-crypto"
	b58 "github.com/mr-tron/base58/base58"
	mh "github.com/multiformats/go-multihash"
)

// MaxInlineKeyLength is the maximum length a key can be for it to be inlined in
// the peer ID.
//
// * When `len(pubKey.Bytes()) <= MaxInlineKeyLength`, the peer ID is the
//   identity multihash hash of the public key.
// * When `len(pubKey.Bytes()) > MaxInlineKeyLength`, the peer ID is the
//   sha2-256 multihash of the public key.
const MaxInlineKeyLength = 42

// ID is a peer identifier.
type ID string

// Pretty returns a b58-encoded string of the ID
func (id ID) Pretty() string {
	return IDB58Encode(id)
}

// String prints out the peer.
func (id ID) String() string {
	pid := id.Pretty()

	// All sha256 nodes start with Qm
	// We can skip the Qm to make the peer.ID more useful
	if strings.HasPrefix(pid, "Qm") {
		pid = pid[2:]
	}

	return pid
}

// MatchesPrivateKey tests whether this ID was derived from sk
func (id ID) MatchesPrivateKey(sk ic.PrivKey) bool {
	return id.MatchesPublicKey(sk.GetPublic())
}

// MatchesPublicKey tests whether this ID was derived from pk
// Returns the expected ID as well.
func (id ID) MatchesPublicKey(pk ic.PubKey) bool {
	oid, err := IDFromPublicKey(pk)
	if err != nil {
		return false
	}

	return oid == id
}

// ExtractPublicKey attempts to extract the public key from an ID
//
// This method returns nil, nil if the peer ID looks valid but it can't extract
// the public key.
func (id ID) ExtractPublicKey() (ic.PubKey, error) {
	decoded, err := mh.Decode([]byte(id))
	if err != nil {
		return nil, err
	}
	if decoded.Code != mh.ID {
		return nil, nil
	}
	pk, err := ic.UnmarshalPublicKey(decoded.Digest)
	if err != nil {
		return nil, err
	}
	return pk, nil
}

// IDFromString cast a string to ID type, and validate
// the id to make sure it is a multihash.
func IDFromString(s string) (ID, error) {
	if _, err := mh.Cast([]byte(s)); err != nil {
		return ID(""), err
	}
	return ID(s), nil
}

// IDFromBytes cast a string to ID type, and validate
// the id to make sure it is a multihash.
func IDFromBytes(b []byte) (ID, error) {
	if _, err := mh.Cast(b); err != nil {
		return ID(""), err
	}

	return ID(b), nil
}

// IDB58Decode returns a b58-decoded Peer
func IDB58Decode(s string) (ID, error) {
	m, err := mh.FromB58String(s)
	if err != nil {
		return "", err
	}
	return ID(m), err
}

// IDB58Encode returns b58-encoded string
func IDB58Encode(id ID) string {
	return b58.Encode([]byte(id))
}

// IDHexDecode returns a hex-decoded Peer
func IDHexDecode(s string) (ID, error) {
	m, err := mh.FromHexString(s)
	if err != nil {
		return "", err
	}
	return ID(m), err
}

// IDHexEncode returns hex-encoded string
func IDHexEncode(id ID) string {
	return hex.EncodeToString([]byte(id))
}

// IDFromPublicKey returns the Peer ID corresponding to pk
func IDFromPublicKey(pk ic.PubKey) (ID, error) {
	b, err := pk.Bytes()
	if err != nil {
		return "", err
	}

	var alg uint64 = mh.SHA2_256
	if len(b) <= MaxInlineKeyLength {
		alg = mh.ID
	}
	hash, _ := mh.Sum(b, alg, -1)
	return ID(hash), nil
}

// IDFromPrivateKey returns the Peer ID corresponding to sk
func IDFromPrivateKey(sk ic.PrivKey) (ID, error) {
	return IDFromPublicKey(sk.GetPublic())
}
