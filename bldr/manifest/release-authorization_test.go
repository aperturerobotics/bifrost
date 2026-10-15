package bldr_manifest

import (
	"testing"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/bucket"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/hash"
	"github.com/s4wave/spacewave/net/peer"
)

// TestReleaseAuthorizationVerifiesInGoScript signs a release authorization with
// the production signer and verifies it with the production verifier. Under
// GoScript, a signing hash the build cannot compute fails here, before the
// browser release smoke that boots the signed manifest.
func TestReleaseAuthorizationVerifiesInGoScript(t *testing.T) {
	// Generate the release key and the pin the verifier trusts.
	key, _, err := crypto.GenerateEd25519Key(nil)
	if err != nil {
		t.Fatal(err)
	}
	releasePeer, err := peer.IDFromPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}

	// Build a release manifest ref over a content-addressed root.
	rootHash, err := hash.Sum(hash.RecommendedHashType, []byte("release manifest root"))
	if err != nil {
		t.Fatal(err)
	}
	meta := &ManifestMeta{ManifestId: "core", BuildType: "production", PlatformId: "js/wasm", Rev: 1}
	ref := NewManifestRef(meta, &bucket.ObjectRef{RootRef: block.NewBlockRef(rootHash)})

	// Sign with the production signer, then verify with the production verifier.
	if err := ref.SignReleaseAuthorization(key); err != nil {
		t.Fatalf("sign release authorization: %v", err)
	}
	if err := ref.VerifyReleaseAuthorization([]peer.ID{releasePeer}); err != nil {
		t.Fatalf("verify release authorization: %v", err)
	}
}
