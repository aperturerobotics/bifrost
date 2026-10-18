package sobject

import (
	"sync"

	"github.com/s4wave/spacewave/net/crypto"
)

// maxKeyringGrants bounds the grants a keyring holds. A shared object holds
// one local grant per key epoch, so the bound is reached only by a long epoch
// history; the keyring then starts over.
const maxKeyringGrants = 64

// SOGrantKeyring decrypts the local peer's grants for the states of one shared
// object and remembers each decrypted grant. Decrypting a grant converts the
// peer's key and runs a scalar multiplication, so the snapshots of one shared
// object share a keyring instead of decrypting the same grant for each new
// state. A grant is immutable while held: state copies share it. A separately
// decoded grant is reused only when its bytes match. The keyring holds the
// transform keys of its grants for its lifetime. It is safe for concurrent use.
type SOGrantKeyring struct {
	// sharedObjectID is the shared object the grants are bound to.
	sharedObjectID string
	// mtx guards inners.
	mtx sync.Mutex
	// inners holds each decrypted grant by its encrypted inner data.
	inners map[string]keyringGrant
}

// keyringGrant is a grant with the body its decryption returned.
type keyringGrant struct {
	// grant is the grant that was decrypted.
	grant *SOGrant
	// inner is the decrypted body of grant, which callers never modify.
	inner *SOGrantInner
}

// NewSOGrantKeyring returns a keyring for the grants of sharedObjectID.
func NewSOGrantKeyring(sharedObjectID string) *SOGrantKeyring {
	return &SOGrantKeyring{sharedObjectID: sharedObjectID}
}

// Inner returns the decrypted body of grant, which is addressed to privKey.
// The caller must not modify the result.
func (k *SOGrantKeyring) Inner(grant *SOGrant, privKey crypto.PrivKey) (*SOGrantInner, error) {
	// Reuse the body of a grant with the same bytes.
	key := string(grant.GetInnerData())
	k.mtx.Lock()
	defer k.mtx.Unlock()
	if held, ok := k.inners[key]; ok && held.grant.EqualVT(grant) {
		return held.inner, nil
	}

	// Decrypt the grant and remember its body.
	inner, err := grant.DecryptInnerData(privKey, k.sharedObjectID)
	if err != nil {
		return nil, err
	}
	if k.inners == nil || len(k.inners) >= maxKeyringGrants {
		k.inners = make(map[string]keyringGrant)
	}
	k.inners[key] = keyringGrant{grant: grant, inner: inner}
	return inner, nil
}
