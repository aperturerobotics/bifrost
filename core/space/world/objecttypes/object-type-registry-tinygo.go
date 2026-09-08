//go:build tinygo

package objecttypes

import (
	"context"

	"github.com/s4wave/spacewave/sdk/world/objecttype"
	wizard_resource "github.com/s4wave/spacewave/sdk/world/wizard/resource"
)

// LookupObjectType looks up a TinyGo-supported object type by ID.
func LookupObjectType(ctx context.Context, typeID string) (objecttype.ObjectType, error) {
	if objectType := compiledObjectTypes[typeID]; objectType != nil {
		return objectType, nil
	}
	return wizard_resource.LookupWizardObjectType(ctx, typeID)
}
