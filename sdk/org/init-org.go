package s4wave_org

import (
	"context"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/world"
	world_types "github.com/s4wave/spacewave/db/world/types"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
)

// OrgObjectKey is the default organization object key within the SO world.
const OrgObjectKey = "org/state"

// OrgRoleOwner is the display role for organization owners.
const OrgRoleOwner = "Owner"

// OrgRoleMember is the display role for organization members.
const OrgRoleMember = "Member"

// ApplyWorldOp applies the init organization operation.
func (o *InitOrganizationOp) ApplyWorldOp(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	sender peer.ID,
) (bool, error) {
	// Resolve the organization key, using the shared object default when omitted.
	objKey := o.GetOrgObjectKey()
	if objKey == "" {
		objKey = OrgObjectKey
	}

	// Prepare the initial organization body with its creator as owner.
	state := &OrgState{
		DisplayName: o.GetDisplayName(),
		CreatedAt:   o.GetTimestamp(),
		Members: []*OrgMemberInfo{{
			AccountId:   o.GetCreatorAccountId(),
			DisplayRole: OrgRoleOwner,
			JoinedAt:    o.GetTimestamp(),
		}},
	}

	// Create the organization World object with its initial body.
	{
		createdObject, _, err := world.CreateWorldObject(ctx, ws, objKey, func(bcs *block.Cursor) error {
			bcs.SetBlock(state, true)
			return nil
		})
		world.ReleaseObjectState(createdObject)
		if err != nil {
			return true, err
		}
	}

	// Register the organization type on the new World object.
	if err := world_types.SetObjectType(ctx, ws, objKey, OrgStateTypeID); err != nil {
		return true, err
	}

	return false, nil
}

var _ world.Operation = (*InitOrganizationOp)(nil)
