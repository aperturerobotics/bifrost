package quad

import (
	"github.com/s4wave/spacewave/db/block"
)

// IsEmpty checks if the graph quad is empty.
// Considered empty if subject, predicate, or object fields are empty.
func (o *Quad) IsEmpty() bool {
	return o.GetSubject() == "" ||
		o.GetPredicate() == "" ||
		o.GetObj() == ""
}

// Clone clones the graph quad.
func (o *Quad) Clone() *Quad {
	if o == nil {
		return nil
	}
	return &Quad{
		Subject:   o.Subject,
		Predicate: o.Predicate,
		Obj:       o.Obj,
		Label:     o.Label,
	}
}

// _ is a type assertion
var _ block.Block = (*Quad)(nil)
