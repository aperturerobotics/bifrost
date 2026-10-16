package world

import (
	"context"

	"github.com/aperturerobotics/common/typeid"
	"github.com/pkg/errors"
)

// Schema resolves World operations by their generated type IDs.
//
// A Schema is built from the []typeid.Type lists that protoc-gen-go-typeid
// generates for each .proto file declaring (typeid.id) options.
type Schema map[string]func() any

// NewSchema builds a Schema from generated type lists.
//
// NewSchema panics if two entries declare the same type ID: the lists are
// fixed at build time, so a collision is a programming error.
func NewSchema(lists ...[]typeid.Type) Schema {
	s := make(Schema)
	for _, list := range lists {
		for _, t := range list {
			if _, ok := s[t.ID]; ok {
				panic("world: type id declared twice: " + t.ID)
			}
			s[t.ID] = t.New
		}
	}
	return s
}

// LookupOp constructs the operation with the type ID.
//
// Returns nil, nil if the schema does not declare the type ID.
func (s Schema) LookupOp(_ context.Context, opTypeID string) (Operation, error) {
	// Other lookups may declare a type ID absent here.
	ctor, ok := s[opTypeID]
	if !ok {
		return nil, nil
	}

	// A schema may also declare types that are not operations.
	op, ok := ctor().(Operation)
	if !ok {
		return nil, errors.Errorf("type id %s is not an operation", opTypeID)
	}
	return op, nil
}

// _ is a type assertion
var _ LookupOp = Schema(nil).LookupOp
