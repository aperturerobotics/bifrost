package forge_lib_kvtx

import (
	"context"

	"github.com/s4wave/spacewave/db/block"
	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	forge_target "github.com/s4wave/spacewave/forge/target"
	forge_value "github.com/s4wave/spacewave/forge/value"
)

// FetchConfigInput fetches the ConfigInput block from the Value.
func FetchConfigInput(
	ctx context.Context,
	handle forge_target.ExecControllerHandle,
	val *forge_value.Value,
) (*ConfigInput, error) {
	// Resolve the configuration value to its bucket reference.
	objRef, err := val.ToBucketRef()
	if err != nil {
		return nil, err
	}

	// Decode the configuration block from the referenced bucket.
	var confInput *ConfigInput
	err = handle.AccessStorage(ctx, objRef, func(bls *bucket_lookup.Cursor) error {
		_, bcs := bls.BuildTransaction(nil)
		var berr error
		confInput, berr = UnmarshalConfigInput(ctx, bcs)
		return berr
	})
	return confInput, err
}

// UnmarshalConfigInput unmarshals a config input block from the cursor.
func UnmarshalConfigInput(ctx context.Context, bcs *block.Cursor) (*ConfigInput, error) {
	return block.UnmarshalBlock[ConfigInput](ctx, bcs)
}

// _ is a type assertion
var _ block.Block = (*ConfigInput)(nil)
