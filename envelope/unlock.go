package envelope

import (
	"bytes"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/util/scrub"
	"github.com/cloudflare/circl/group"
	"github.com/cloudflare/circl/secretsharing"
	"github.com/libp2p/go-libp2p/core/crypto"
	"golang.org/x/crypto/chacha20poly1305"
)

// UnlockEnvelope attempts to decrypt an Envelope using the provided private keys.
//
// Returns (payload, result, nil) on success.
// Returns (nil, result, nil) when not enough shares are available (result has progress).
// Returns (nil, nil, err) on invalid envelope or other errors.
func UnlockEnvelope(
	context string,
	env *Envelope,
	privKeys []crypto.PrivKey,
) ([]byte, *EnvelopeUnlockResult, error) {
	if len(env.GetGrants()) == 0 {
		return nil, nil, ErrNoGrants
	}
	if len(env.GetKeypairs()) == 0 {
		return nil, nil, ErrNoKeypairs
	}

	// Verify context hash.
	expected := hashContext(context)
	if !bytes.Equal(env.GetContextHash(), expected) {
		return nil, nil, ErrContextMismatch
	}

	// Match private keys to envelope keypairs.
	matched, err := matchPrivKeys(env, privKeys)
	if err != nil {
		return nil, nil, err
	}

	threshold := env.GetThreshold()
	sharesNeeded := threshold + 1
	envelopeID := env.GetEnvelopeId()

	// Decrypt grants and collect shares.
	var collected []secretsharing.Share
	var unlockedIndexes []uint32

	for gi, grant := range env.GetGrants() {
		kpIndexes := grant.GetKeypairIndexes()
		ciphertexts := grant.GetCiphertexts()
		if len(kpIndexes) != len(ciphertexts) {
			continue
		}

		encCtx := buildGrantEncContext(envelopeID, context, gi)

		var innerData []byte
		for ci, kpIdx := range kpIndexes {
			priv, ok := matched[int(kpIdx)]
			if !ok {
				continue
			}
			dec, err := peer.DecryptWithPrivKey(priv, encCtx, ciphertexts[ci])
			if err != nil {
				continue
			}
			innerData = dec
			break
		}
		if innerData == nil {
			continue
		}
		defer scrub.Scrub(innerData)

		inner := &EnvelopeGrantInner{}
		if err := inner.UnmarshalVT(innerData); err != nil {
			continue
		}

		unlockedIndexes = append(unlockedIndexes, uint32(gi))

		g := group.Ristretto255
		for _, s := range inner.GetShares() {
			id := g.NewScalar()
			if err := id.UnmarshalBinary(s.GetId()); err != nil {
				continue
			}
			val := g.NewScalar()
			if err := val.UnmarshalBinary(s.GetValue()); err != nil {
				continue
			}
			collected = append(collected, secretsharing.Share{ID: id, Value: val})
		}
	}

	result := &EnvelopeUnlockResult{
		SharesAvailable:     uint32(len(collected)),
		SharesNeeded:        sharesNeeded,
		UnlockedGrantIndexes: unlockedIndexes,
	}

	if uint32(len(collected)) < sharesNeeded {
		return nil, result, nil
	}

	// Recover the secret scalar via CIRCL Lagrange interpolation.
	recovered, err := secretsharing.Recover(uint(threshold), collected)
	if err != nil {
		return nil, nil, err
	}
	scalarBytes, err := recovered.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}
	defer scrub.Scrub(scalarBytes)

	// Derive encryption key and decrypt payload.
	encKey := deriveEncKeyFromScalar(scalarBytes, envelopeID, context)
	defer scrub.Scrub(encKey[:])

	aead, err := chacha20poly1305.NewX(encKey[:])
	if err != nil {
		return nil, nil, err
	}

	ct := env.GetCiphertext()
	if len(ct) < aead.NonceSize() {
		return nil, nil, ErrDecryptionFailed
	}

	nonce := ct[:aead.NonceSize()]
	payload, err := aead.Open(nil, nonce, ct[aead.NonceSize():], nil)
	if err != nil {
		return nil, nil, ErrDecryptionFailed
	}

	result.Success = true
	return payload, result, nil
}
