package blockcrypt

import (
	"github.com/paralin/kcp-go-lite"
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
		return kcp.NewSM4BlockCrypt(pass[:16])
	case BlockCrypt_BlockCrypt_TEA16:
		return kcp.NewTEABlockCrypt(pass[:16])
	case BlockCrypt_BlockCrypt_XOR:
		return kcp.NewSimpleXORBlockCrypt(pass)
	case BlockCrypt_BlockCrypt_NONE:
		return kcp.NewNoneBlockCrypt(pass)
	case BlockCrypt_BlockCrypt_AES128:
		return kcp.NewAESBlockCrypt(pass[:16])
	case BlockCrypt_BlockCrypt_AES192:
		return kcp.NewAESBlockCrypt(pass[:24])
	case BlockCrypt_BlockCrypt_BLOWFISH:
		return kcp.NewBlowfishBlockCrypt(pass)
	case BlockCrypt_BlockCrypt_TWOFISH:
		return kcp.NewTwofishBlockCrypt(pass)
	case BlockCrypt_BlockCrypt_CAST5:
		return kcp.NewCast5BlockCrypt(pass[:16])
	case BlockCrypt_BlockCrypt_3DES:
		return kcp.NewTripleDESBlockCrypt(pass[:24])
	case BlockCrypt_BlockCrypt_XTEA:
		return kcp.NewXTEABlockCrypt(pass[:16])
	case BlockCrypt_BlockCrypt_SALSA20:
		return kcp.NewSalsa20BlockCrypt(pass)
	case BlockCrypt_BlockCrypt_UNKNOWN:
		fallthrough
	case BlockCrypt_BlockCrypt_AES256:
		return kcp.NewAESBlockCrypt(pass)
	default:
		return nil, errors.Errorf("unrecognized blockcrypt type: %s", crypt.String())
	}
}
