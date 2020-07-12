package keyfile

import (
	"io/ioutil"
	"os"

	"github.com/aperturerobotics/bifrost/keypem"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/sirupsen/logrus"
)

// OpenOrWritePrivKey opens or generates a private key at a path.
// Uses PEM format, defaults for GeneratePrivKey that Bifrost uses.
// Note: this will be the latest form we use, and is backwards compat.
// Ex: both RSA and Ed25518 are key types
// May return a private key + an error.
func OpenOrWritePrivKey(le *logrus.Entry, privKeyPath string) (crypto.PrivKey, error) {
	var privKey crypto.PrivKey
	var err error
	if _, err := os.Stat(privKeyPath); err != nil {
		if os.IsNotExist(err) {
			if le != nil {
				le.Debug("generating priv key")
			}
			privKey, _, err = keypem.GeneratePrivKey()
			if err != nil {
				return privKey, err
			}
			dat, err := keypem.MarshalPrivKeyPem(privKey)
			if err != nil {
				return privKey, err
			}
			if err := ioutil.WriteFile(privKeyPath, dat, 0600); err != nil {
				return privKey, err
			}
			if le != nil {
				le.Debug("wrote private key")
			}
		}
	} else {
		dat, err := ioutil.ReadFile(privKeyPath)
		if err != nil {
			return privKey, err
		}
		privKey, err = keypem.ParsePrivKeyPem(dat)
		if err != nil {
			return privKey, err
		}
	}
	return privKey, err
}
