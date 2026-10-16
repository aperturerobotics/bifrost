package s4wave_org

import (
	"context"

	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/block"
)

// OrganizationTypeID is the type identifier for organization objects.
const OrganizationTypeID = "spacewave/organization"

// OrgBodyType is the SharedObject body type for organization SOs.
const OrgBodyType = "organization"

// NewOrgSharedObjectMeta constructs a SharedObjectMeta for an org SO.
// The displayName is stored in BodyMeta for use by org list scans.
func NewOrgSharedObjectMeta(displayName string) *sobject.SharedObjectMeta {
	return &sobject.SharedObjectMeta{
		BodyType: OrgBodyType,
		BodyMeta: []byte(displayName),
	}
}

// OrgDisplayNameFromMeta extracts the display name from an org SO's BodyMeta.
func OrgDisplayNameFromMeta(meta *sobject.SharedObjectMeta) string {
	if meta == nil {
		return ""
	}
	return string(meta.GetBodyMeta())
}

// UnmarshalOrgState unmarshals an org state from a cursor.
func UnmarshalOrgState(ctx context.Context, bcs *block.Cursor) (*OrgState, error) {
	return block.UnmarshalBlock[OrgState](ctx, bcs)
}

// Validate performs cursory checks on the OrgState.
func (o *OrgState) Validate() error {
	return nil
}

var _ block.Block = (*OrgState)(nil)
