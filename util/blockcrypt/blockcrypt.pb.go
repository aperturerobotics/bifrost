// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/aperturerobotics/bifrost/util/blockcrypt/blockcrypt.proto

package blockcrypt

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// BlockCrypt sets the type of block crypto to use.
type BlockCrypt int32

const (
	// BlockCrypt_UNKNOWN defaults to BlockCrypt_AES256
	BlockCrypt_BlockCrypt_UNKNOWN BlockCrypt = 0
	// BlockCrypt_AES256 is AES 256-bit block encryption.
	BlockCrypt_BlockCrypt_AES256 BlockCrypt = 1
	// BlockCrypt_AES128 is AES 128-bit block encryption.
	BlockCrypt_BlockCrypt_AES128 BlockCrypt = 2
	// BlockCrypt_AES192 is AES 192-bit block encryption.
	BlockCrypt_BlockCrypt_AES192 BlockCrypt = 3
	// BlockCrypt_SM4_16 is SM4 16-bit block encryption.
	BlockCrypt_BlockCrypt_SM4_16 BlockCrypt = 4
	// BlockCrypt_TEA16 is 16-bit TEA block encryption.
	BlockCrypt_BlockCrypt_TEA16 BlockCrypt = 5
	// BlockCrypt_XOR is simple XOR block encryption.
	BlockCrypt_BlockCrypt_XOR BlockCrypt = 6
	// BlockCrypt_NONE is unencrypted.
	BlockCrypt_BlockCrypt_NONE BlockCrypt = 7
	// BlockCrypt_BLOWFISH is blowfish 32-bit block encryption.
	BlockCrypt_BlockCrypt_BLOWFISH BlockCrypt = 8
	// BlockCrypt_TWOFISH is twofish 32-bit block encryption.
	BlockCrypt_BlockCrypt_TWOFISH BlockCrypt = 9
	// BlockCrypt_CAST5 is cast5 16bit block encryption.
	BlockCrypt_BlockCrypt_CAST5 BlockCrypt = 10
	// BlockCrypt_3DES is 3des 24-bit block encryption.
	BlockCrypt_BlockCrypt_3DES BlockCrypt = 11
	// BlockCrypt_XTEA is xtea 16-bit block encryption.
	BlockCrypt_BlockCrypt_XTEA BlockCrypt = 12
	// BlockCrypt_SALSA20 is salsa20 32-bit block encryption.
	BlockCrypt_BlockCrypt_SALSA20 BlockCrypt = 13
)

var BlockCrypt_name = map[int32]string{
	0:  "BlockCrypt_UNKNOWN",
	1:  "BlockCrypt_AES256",
	2:  "BlockCrypt_AES128",
	3:  "BlockCrypt_AES192",
	4:  "BlockCrypt_SM4_16",
	5:  "BlockCrypt_TEA16",
	6:  "BlockCrypt_XOR",
	7:  "BlockCrypt_NONE",
	8:  "BlockCrypt_BLOWFISH",
	9:  "BlockCrypt_TWOFISH",
	10: "BlockCrypt_CAST5",
	11: "BlockCrypt_3DES",
	12: "BlockCrypt_XTEA",
	13: "BlockCrypt_SALSA20",
}

var BlockCrypt_value = map[string]int32{
	"BlockCrypt_UNKNOWN":  0,
	"BlockCrypt_AES256":   1,
	"BlockCrypt_AES128":   2,
	"BlockCrypt_AES192":   3,
	"BlockCrypt_SM4_16":   4,
	"BlockCrypt_TEA16":    5,
	"BlockCrypt_XOR":      6,
	"BlockCrypt_NONE":     7,
	"BlockCrypt_BLOWFISH": 8,
	"BlockCrypt_TWOFISH":  9,
	"BlockCrypt_CAST5":    10,
	"BlockCrypt_3DES":     11,
	"BlockCrypt_XTEA":     12,
	"BlockCrypt_SALSA20":  13,
}

func (x BlockCrypt) String() string {
	return proto.EnumName(BlockCrypt_name, int32(x))
}

func (BlockCrypt) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_1b0289c9e855eb5b, []int{0}
}

func init() {
	proto.RegisterEnum("blockcrypt.BlockCrypt", BlockCrypt_name, BlockCrypt_value)
}

func init() {
	proto.RegisterFile("github.com/aperturerobotics/bifrost/util/blockcrypt/blockcrypt.proto", fileDescriptor_1b0289c9e855eb5b)
}

var fileDescriptor_1b0289c9e855eb5b = []byte{
	// 250 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0xd0, 0x4b, 0x4e, 0xc3, 0x30,
	0x10, 0x06, 0x60, 0x28, 0x50, 0x60, 0x78, 0x0d, 0x53, 0x1e, 0x77, 0x60, 0xd1, 0x90, 0x94, 0x46,
	0xb0, 0x74, 0x5b, 0x23, 0x10, 0xc5, 0x96, 0x70, 0x50, 0xba, 0xab, 0x70, 0x54, 0x20, 0xa2, 0xc8,
	0x91, 0xeb, 0x2c, 0xb8, 0x21, 0xc7, 0x42, 0xed, 0x26, 0x96, 0xe9, 0x6e, 0xfe, 0x6f, 0xf1, 0x8f,
	0x66, 0x60, 0xf4, 0x51, 0xba, 0xcf, 0x5a, 0x77, 0x0b, 0xf3, 0x1d, 0xbd, 0x55, 0x33, 0xeb, 0x6a,
	0x3b, 0xb3, 0x46, 0x1b, 0x57, 0x16, 0x8b, 0x48, 0x97, 0xef, 0xd6, 0x2c, 0x5c, 0x54, 0xbb, 0x72,
	0x1e, 0xe9, 0xb9, 0x29, 0xbe, 0x0a, 0xfb, 0x53, 0x39, 0x6f, 0xec, 0x56, 0xd6, 0x38, 0x43, 0xd0,
	0xc8, 0xd5, 0x6f, 0x0b, 0x60, 0xb0, 0x8c, 0xc3, 0x65, 0xa4, 0x0b, 0xa0, 0x26, 0x4d, 0x5f, 0xc5,
	0x93, 0x90, 0xb9, 0xc0, 0x0d, 0x3a, 0x87, 0x53, 0xcf, 0x19, 0x57, 0x49, 0x3f, 0xc5, 0xcd, 0xff,
	0x1c, 0x27, 0xb7, 0xd8, 0x5a, 0xc3, 0x77, 0x09, 0x6e, 0x05, 0xac, 0x9e, 0x6f, 0xa6, 0x71, 0x8a,
	0xdb, 0x74, 0x06, 0xe8, 0x71, 0xc6, 0x59, 0x9c, 0xe2, 0x0e, 0x11, 0x1c, 0x7b, 0x3a, 0x91, 0x2f,
	0xd8, 0xa6, 0x0e, 0x9c, 0x78, 0x26, 0xa4, 0xe0, 0xb8, 0x4b, 0x97, 0xd0, 0xf1, 0x70, 0x30, 0x96,
	0xf9, 0xfd, 0xa3, 0x7a, 0xc0, 0xbd, 0xe0, 0x96, 0x2c, 0x97, 0x2b, 0xdf, 0x0f, 0xf6, 0x0d, 0x99,
	0xca, 0xfa, 0x08, 0x41, 0x77, 0x6f, 0xc4, 0x15, 0x1e, 0x04, 0x38, 0xc9, 0x38, 0xc3, 0xc3, 0xa0,
	0x57, 0xb1, 0xb1, 0x62, 0xc9, 0x35, 0x1e, 0xe9, 0xf6, 0xea, 0xbb, 0xbd, 0xbf, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xab, 0x7e, 0x8a, 0x79, 0xa5, 0x01, 0x00, 0x00,
}