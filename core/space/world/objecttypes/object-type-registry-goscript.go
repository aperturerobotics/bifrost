//go:build goscript

package objecttypes

import (
	"context"
	"strings"

	"github.com/s4wave/spacewave/sdk/world/objecttype"
	wizard "github.com/s4wave/spacewave/sdk/world/wizard"
	wizard_resource "github.com/s4wave/spacewave/sdk/world/wizard/resource"
)

// LookupObjectType looks up a GoScript-supported object type by ID.
func LookupObjectType(ctx context.Context, typeID string) (objecttype.ObjectType, error) {
	if objectType := compiledObjectTypes[typeID]; objectType != nil {
		return objectType, nil
	}
	if !strings.HasPrefix(typeID, wizard.WizardTypePrefix) {
		return nil, nil
	}
	return wizard_resource.LookupWizardObjectType(ctx, typeID)
}
