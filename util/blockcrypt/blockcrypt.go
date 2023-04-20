package blockcrypt

import (
	kcrypt "github.com/aperturerobotics/bifrost/util/blockcrypt/crypt"
	"github.com/pkg/errors"
)

// BlockCrypt_BlockCrypt_MAX is the maximum value for BlockCrypt.
const BlockCrypt_BlockCrypt_MAX = BlockCrypt_BlockCrypt_SALSA20

// Crypt defines encryption/decryption methods for a given byte slice.
// Notes on implementing: the data to be encrypted contains a builtin
// nonce at the first 16 bytes
type Crypt interface {
	// Encrypt encrypts the whole block in src into dst.
	// Dst and src may point at the same memory.
	Encrypt(dst, src []byte)

	// Decrypt decrypts the whole block in src into dst.
	// Dst and src may point at the same memory.
	Decrypt(dst, src []byte)
}

// BuildBlockCrypt builds block crypt from known types.
func BuildBlockCrypt(crypt BlockCrypt, pass []byte) (Crypt, error) {
	switch crypt {
	case BlockCrypt_BlockCrypt_XOR:
		return kcrypt.NewSimpleXORBlockCrypt(pass)
	case BlockCrypt_BlockCrypt_NONE:
		return kcrypt.NewNoneBlockCrypt(pass)
	case BlockCrypt_BlockCrypt_AES128:
		return kcrypt.NewAESBlockCrypt(pass[:16])
	case BlockCrypt_BlockCrypt_AES192:
		return kcrypt.NewAESBlockCrypt(pass[:24])
	case BlockCrypt_BlockCrypt_3DES:
		return kcrypt.NewTripleDESBlockCrypt(pass[:24])
	case BlockCrypt_BlockCrypt_SALSA20:
		return kcrypt.NewSalsa20BlockCrypt(pass)
	case BlockCrypt_BlockCrypt_UNKNOWN:
		fallthrough
	case BlockCrypt_BlockCrypt_AES256:
		return kcrypt.NewAESBlockCrypt(pass)
	default:
		return nil, errors.Errorf("unrecognized blockcrypt type: %s", crypt.String())
	}
}
