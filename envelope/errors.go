package envelope

import "github.com/pkg/errors"

var (
	ErrNoGrants            = errors.New("envelope has no grants")
	ErrNoKeypairs          = errors.New("envelope has no keypairs")
	ErrInvalidThreshold    = errors.New("invalid threshold configuration")
	ErrContextMismatch     = errors.New("envelope context does not match expected context")
	ErrInsufficientShares  = errors.New("insufficient shares to unlock envelope")
	ErrInvalidShareData    = errors.New("invalid share data")
	ErrInvalidKeypairIndex = errors.New("keypair index out of range")
	ErrEmptyPayload        = errors.New("payload is empty")
	ErrDecryptionFailed    = errors.New("envelope decryption failed")
)
