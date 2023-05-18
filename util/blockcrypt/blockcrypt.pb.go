// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0-devel
// 	protoc        v3.21.9
// source: github.com/aperturerobotics/bifrost/util/blockcrypt/blockcrypt.proto

package blockcrypt

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// BlockCrypt sets the type of block crypto to use.
type BlockCrypt int32

const (
	// BlockCrypt_UNKNOWN defaults to BlockCrypt_AES256
	BlockCrypt_BlockCrypt_UNKNOWN BlockCrypt = 0
	// BlockCrypt_NONE is unencrypted.
	BlockCrypt_BlockCrypt_NONE BlockCrypt = 1
	// BlockCrypt_AES256 is AES 256-bit block encryption.
	BlockCrypt_BlockCrypt_AES256 BlockCrypt = 2
	// BlockCrypt_AES128 is AES 128-bit block encryption.
	BlockCrypt_BlockCrypt_AES128 BlockCrypt = 3
	// BlockCrypt_AES192 is AES 192-bit block encryption.
	BlockCrypt_BlockCrypt_AES192 BlockCrypt = 4
	// BlockCrypt_XOR is simple XOR block encryption.
	BlockCrypt_BlockCrypt_XOR BlockCrypt = 6
	// BlockCrypt_3DES is 3des 24-bit block encryption.
	BlockCrypt_BlockCrypt_3DES BlockCrypt = 7
	// BlockCrypt_SALSA20 is salsa20 32-bit block encryption.
	BlockCrypt_BlockCrypt_SALSA20 BlockCrypt = 8
)

// Enum value maps for BlockCrypt.
var (
	BlockCrypt_name = map[int32]string{
		0: "BlockCrypt_UNKNOWN",
		1: "BlockCrypt_NONE",
		2: "BlockCrypt_AES256",
		3: "BlockCrypt_AES128",
		4: "BlockCrypt_AES192",
		6: "BlockCrypt_XOR",
		7: "BlockCrypt_3DES",
		8: "BlockCrypt_SALSA20",
	}
	BlockCrypt_value = map[string]int32{
		"BlockCrypt_UNKNOWN": 0,
		"BlockCrypt_NONE":    1,
		"BlockCrypt_AES256":  2,
		"BlockCrypt_AES128":  3,
		"BlockCrypt_AES192":  4,
		"BlockCrypt_XOR":     6,
		"BlockCrypt_3DES":    7,
		"BlockCrypt_SALSA20": 8,
	}
)

func (x BlockCrypt) Enum() *BlockCrypt {
	p := new(BlockCrypt)
	*p = x
	return p
}

func (x BlockCrypt) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (BlockCrypt) Descriptor() protoreflect.EnumDescriptor {
	return file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_enumTypes[0].Descriptor()
}

func (BlockCrypt) Type() protoreflect.EnumType {
	return &file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_enumTypes[0]
}

func (x BlockCrypt) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use BlockCrypt.Descriptor instead.
func (BlockCrypt) EnumDescriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_rawDescGZIP(), []int{0}
}

var File_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_rawDesc = []byte{
	0x0a, 0x44, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x75, 0x74, 0x69, 0x6c, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x63, 0x72, 0x79, 0x70, 0x74, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x72, 0x79, 0x70, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x72, 0x79,
	0x70, 0x74, 0x2a, 0xbf, 0x01, 0x0a, 0x0a, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x72, 0x79, 0x70,
	0x74, 0x12, 0x16, 0x0a, 0x12, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x72, 0x79, 0x70, 0x74, 0x5f,
	0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x13, 0x0a, 0x0f, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x43, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x4e, 0x4f, 0x4e, 0x45, 0x10, 0x01, 0x12, 0x15,
	0x0a, 0x11, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x41, 0x45, 0x53,
	0x32, 0x35, 0x36, 0x10, 0x02, 0x12, 0x15, 0x0a, 0x11, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x72,
	0x79, 0x70, 0x74, 0x5f, 0x41, 0x45, 0x53, 0x31, 0x32, 0x38, 0x10, 0x03, 0x12, 0x15, 0x0a, 0x11,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x41, 0x45, 0x53, 0x31, 0x39,
	0x32, 0x10, 0x04, 0x12, 0x12, 0x0a, 0x0e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x72, 0x79, 0x70,
	0x74, 0x5f, 0x58, 0x4f, 0x52, 0x10, 0x06, 0x12, 0x13, 0x0a, 0x0f, 0x42, 0x6c, 0x6f, 0x63, 0x6b,
	0x43, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x33, 0x44, 0x45, 0x53, 0x10, 0x07, 0x12, 0x16, 0x0a, 0x12,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x72, 0x79, 0x70, 0x74, 0x5f, 0x53, 0x41, 0x4c, 0x53, 0x41,
	0x32, 0x30, 0x10, 0x08, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_rawDescData = file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_goTypes = []interface{}{
	(BlockCrypt)(0), // 0: blockcrypt.BlockCrypt
}
var file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_init() }
func file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_init() {
	if File_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_depIdxs,
		EnumInfos:         file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_enumTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto = out.File
	file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_util_blockcrypt_blockcrypt_proto_depIdxs = nil
}
