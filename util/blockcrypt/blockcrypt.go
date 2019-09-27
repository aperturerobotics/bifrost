package blockcrypt

import (
	kcrypt "github.com/aperturerobotics/bifrost/util/blockcrypt/crypt"
	"github.com/pkg/errors"
)

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
	case BlockCrypt_BlockCrypt_SM4_16:
		return kcrypt.NewSM4BlockCrypt(pass[:16])
	case BlockCrypt_BlockCrypt_TEA16:
		return kcrypt.NewTEABlockCrypt(pass[:16])
	case BlockCrypt_BlockCrypt_XOR:
		return kcrypt.NewSimpleXORBlockCrypt(pass)
	case BlockCrypt_BlockCrypt_NONE:
		return kcrypt.NewNoneBlockCrypt(pass)
	case BlockCrypt_BlockCrypt_AES128:
		return kcrypt.NewAESBlockCrypt(pass[:16])
	case BlockCrypt_BlockCrypt_AES192:
		return kcrypt.NewAESBlockCrypt(pass[:24])
	case BlockCrypt_BlockCrypt_BLOWFISH:
		return kcrypt.NewBlowfishBlockCrypt(pass)
	case BlockCrypt_BlockCrypt_TWOFISH:
		return kcrypt.NewTwofishBlockCrypt(pass)
	case BlockCrypt_BlockCrypt_CAST5:
		return kcrypt.NewCast5BlockCrypt(pass[:16])
	case BlockCrypt_BlockCrypt_3DES:
		return kcrypt.NewTripleDESBlockCrypt(pass[:24])
	case BlockCrypt_BlockCrypt_XTEA:
		return kcrypt.NewXTEABlockCrypt(pass[:16])
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
