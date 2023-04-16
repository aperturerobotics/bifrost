package hash

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"hash"
	"strconv"

	b58 "github.com/mr-tron/base58/base58"
	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
	"github.com/zeebo/blake3"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"
)

// ErrHashMismatch is returned when hashes mismatch.
var ErrHashMismatch = errors.New("hash mismatch")

// ErrHashTypeUnknown is returned when the hash type is unknown.
var ErrHashTypeUnknown = errors.New("unknown hash type")

// SupportedHashTypes is the list of built-in hash types.
var SupportedHashTypes = []HashType{
	HashType_HashType_SHA256,
	HashType_HashType_SHA1,
	HashType_HashType_BLAKE3,
}

// RecommendedHashType is the hash type recommended to use.
// Note: not guaranteed to stay the same between Bifrost versions.
const RecommendedHashType = HashType_HashType_BLAKE3

// UnmarshalHashTypeFastJSON unmarshals a HashType from FastJSON value.
func UnmarshalHashTypeFastJSON(val *fastjson.Value) (HashType, error) {
	if val == nil {
		return 0, nil
	}
	switch val.Type() {
	case fastjson.TypeString:
		valStr := string(val.GetStringBytes())
		if len(valStr) == 0 {
			return 0, nil
		}
		val, ok := HashType_value[valStr]
		if !ok {
			return 0, errors.Wrap(ErrHashTypeUnknown, valStr)
		}
		return HashType(val), nil
	case fastjson.TypeNumber:
		ht := HashType(val.GetInt())
		if err := ht.Validate(); err != nil {
			return 0, err
		}
		return ht, nil
	default:
		return 0, errors.Errorf("unexpected json type for hash type: %v", val.Type().String())
	}
}

// UnmarshalHashJSON unmarshals a Hash from JSON.
func UnmarshalHashJSON(data []byte) (*Hash, error) {
	val, err := fastjson.Parse(string(data))
	if err != nil {
		return nil, err
	}
	return UnmarshalHashFastJSON(val)
}

// UnmarshalHashFastJSON unmarshals a Hash from FastJSON value.
func UnmarshalHashFastJSON(val *fastjson.Value) (*Hash, error) {
	out := &Hash{}
	if val == nil {
		return out, nil
	}
	if err := out.UnmarshalFastJSON(val); err != nil {
		return nil, err
	}
	return out, nil
}

// IsNil checks if the object is nil.
func (h *Hash) IsNil() bool {
	return h == nil
}

// IsEmpty checks if the hash is empty.
func (h *Hash) IsEmpty() bool {
	return h.GetHashType() == 0 || len(h.GetHash()) == 0
}

// Clone clones the hash object.
func (h *Hash) Clone() *Hash {
	if h == nil {
		return nil
	}

	return &Hash{
		HashType: h.GetHashType(),
		Hash:     slices.Clone(h.GetHash()),
	}
}

// Validate validates the hash type.
func (h HashType) Validate() error {
	switch h {
	case HashType_HashType_UNKNOWN:
		return nil
	case HashType_HashType_SHA256:
		return nil
	case HashType_HashType_SHA1:
		return nil
	case HashType_HashType_BLAKE3:
		return nil
	default:
		return errors.Errorf("hash type unknown: %v", h.String())
	}
}

// GetHashLen returns the hash length.
func (h HashType) GetHashLen() int {
	switch h {
	case HashType_HashType_SHA256:
		return sha256.Size
	case HashType_HashType_SHA1:
		return sha1.Size
	case HashType_HashType_BLAKE3:
		return 32
	}
	return 0
}

// Sum takes the sum with the hash type.
func (h HashType) Sum(data []byte) ([]byte, error) {
	switch h {
	case HashType_HashType_SHA256:
		h := sha256.Sum256(data)
		return h[:], nil
	case HashType_HashType_SHA1:
		h := sha1.Sum(data)
		return h[:], nil
	case HashType_HashType_BLAKE3:
		h := blake3.Sum256(data)
		return h[:], nil
	default:
		return nil, errors.Errorf("hash type unknown: %v", h.String())
	}
}

// CompareHash compares two hashes.
func (h *Hash) CompareHash(other *Hash) bool {
	if other == nil && h == nil {
		return true
	}
	if h == nil || other == nil {
		return false
	}
	if h.GetHashType() != other.GetHashType() {
		return false
	}
	if len(h.GetHash()) != len(other.GetHash()) {
		return false
	}
	if !bytes.Equal(h.GetHash(), other.GetHash()) {
		return false
	}
	return true
}

// BuildHasher builds the hasher for the hash type.
func (h HashType) BuildHasher() (hash.Hash, error) {
	switch h {
	case HashType_HashType_SHA256:
		return sha256.New(), nil
	case HashType_HashType_SHA1:
		return sha1.New(), nil
	case HashType_HashType_BLAKE3:
		return blake3.New(), nil
	default:
		return nil, errors.Errorf("hash type unknown: %v", h.String())
	}
}

// VerifyData verifies data against the sum.
// Returns the hash of the data, hash type, and error
// Returns an error if failed to validate.
func (h *Hash) VerifyData(data []byte) ([]byte, error) {
	hash, err := h.GetHashType().Sum(data)
	if err != nil {
		return nil, err
	}
	if len(hash) != len(h.GetHash()) {
		return hash, ErrHashMismatch
	}
	if !bytes.Equal(hash, h.GetHash()) {
		return hash, ErrHashMismatch
	}
	return hash, nil
}

// NewHash constructs a new hash object.
func NewHash(ht HashType, h []byte) *Hash {
	return &Hash{HashType: ht, Hash: h}
}

// Sum constructs a hash type by summing an object.
func Sum(ht HashType, data []byte) (*Hash, error) {
	h, err := ht.Sum(data)
	if err != nil {
		return nil, err
	}
	return NewHash(ht, h), nil
}

// Validate validates the hash.
func (h *Hash) Validate() error {
	if err := h.GetHashType().Validate(); err != nil {
		return err
	}
	if ehl := h.GetHashType().GetHashLen(); len(h.GetHash()) != ehl {
		return errors.Errorf("expected hash length %d != %d", ehl, len(h.GetHash()))
	}
	return nil
}

// MarshalString marshals the hash to a string.
func (h *Hash) MarshalString() string {
	if h == nil {
		return ""
	}
	dat, err := proto.Marshal(h)
	if err != nil {
		return ""
	}
	return b58.Encode(dat)
}

// MarshalDigest marshals the hash to a binary slice.
func (h *Hash) MarshalDigest() []byte {
	if h == nil {
		return nil
	}
	dat, _ := proto.Marshal(h)
	return dat
}

// MarshalJSON marshals the hash to a JSON string.
// Returns "" if the hash is nil.
func (h *Hash) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(h.MarshalString())), nil
}

// UnmarshalJSON unmarshals the reference from a JSON string or object.
// Also accepts an object (in jsonpb format).
func (h *Hash) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || h == nil {
		return nil
	}
	val, err := fastjson.ParseBytes(data)
	if err != nil {
		return err
	}
	return h.UnmarshalFastJSON(val)
}

// ParseFromB58 parses the object ref from a base58 string.
func (h *Hash) ParseFromB58(ref string) error {
	dat, err := b58.Decode(ref)
	if err != nil {
		return err
	}
	return h.UnmarshalVT(dat)
}

// UnmarshalFastJSON unmarshals the fast json container.
// If the val or object ref are nil, does nothing.
func (h *Hash) UnmarshalFastJSON(val *fastjson.Value) error {
	if val == nil || h == nil {
		return nil
	}
	switch val.Type() {
	case fastjson.TypeString:
		return h.ParseFromB58(string(val.GetStringBytes()))
	case fastjson.TypeObject:

	default:
		return errors.Errorf("unexpected json type for hash: %v", val.Type().String())
	}

	if hashTypeVal := val.Get("hashType"); hashTypeVal != nil {
		var err error
		h.HashType, err = UnmarshalHashTypeFastJSON(hashTypeVal)
		if err != nil {
			return err
		}
	}

	if hashVal := val.Get("hash"); hashVal != nil {
		// expect b58 string
		hashStr, err := hashVal.StringBytes()
		if err != nil {
			return errors.Wrap(err, "hash")
		}
		hashData, err := b58.Decode(string(hashStr))
		if err != nil {
			return errors.Wrap(err, "hash")
		}
		h.Hash = hashData
	}

	return nil
}

var (
	_ json.Marshaler   = ((*Hash)(nil))
	_ json.Unmarshaler = ((*Hash)(nil))
)
