package peer_ssh

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"golang.org/x/crypto/ssh"
)

// NewSigner constructs a new signer from a private key
func NewSigner(privKey crypto.PrivKey) (ssh.Signer, error) {
	cryptoPriv, err := crypto.PrivKeyToStdKey(privKey)
	if err != nil {
		return nil, err
	}
	return ssh.NewSignerFromKey(cryptoPriv)
}
