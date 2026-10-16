package s4wave_apt

import (
	"context"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/world"
	world_types "github.com/s4wave/spacewave/db/world/types"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
)

// NewCreateAptRepositoryOp constructs a new CreateAptRepositoryOp.
func NewCreateAptRepositoryOp(objectKey string, repository *AptRepository) *CreateAptRepositoryOp {
	return &CreateAptRepositoryOp{
		ObjectKey:  objectKey,
		Repository: repository,
	}
}

// Validate performs cursory checks on the op.
func (o *CreateAptRepositoryOp) Validate() error {
	// Validate the repository identity, contents, and initial state.
	if o.GetObjectKey() == "" {
		return world.ErrEmptyObjectKey
	}
	if o.GetRepository() == nil {
		return errors.New("repository is required")
	}
	if err := o.GetRepository().Validate(); err != nil {
		return err
	}
	if state := o.GetRepository().GetState(); state != AptRepositoryState_AptRepositoryState_EMPTY {
		return errors.Wrap(ErrInvalidAptRepositoryInitialState, state.String())
	}
	return nil
}

// ApplyWorldOp applies the operation as a world operation.
func (o *CreateAptRepositoryOp) ApplyWorldOp(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	sender peer.ID,
) (sysErr bool, err error) {
	// Validate the repository creation request before changing the World.
	if err := o.Validate(); err != nil {
		return false, err
	}

	// Create the repository object from a copy of the requested record.
	objectKey := o.GetObjectKey()
	repository := o.GetRepository().CloneVT()
	var createdObject world.ObjectState
	createdObject, _, err = world.CreateWorldObject(ctx, ws, objectKey, func(bcs *block.Cursor) error {
		bcs.SetBlock(repository, true)
		return nil
	})
	world.ReleaseObjectState(createdObject)
	if err != nil {
		return false, err
	}

	// Register the repository object type in the World graph.
	if err := world_types.SetObjectType(ctx, ws, objectKey, AptRepositoryTypeID); err != nil {
		return false, err
	}

	return false, nil
}

// NewAddAptPackageOp constructs a new AddAptPackageOp.
func NewAddAptPackageOp(repositoryKey, packageKey string, aptPackage *AptPackage) *AddAptPackageOp {
	return &AddAptPackageOp{
		RepositoryKey: repositoryKey,
		PackageKey:    packageKey,
		AptPackage:    aptPackage,
	}
}

// Validate performs cursory checks on the op.
func (o *AddAptPackageOp) Validate() error {
	// Require the repository and child record keys and contents.
	if o.GetRepositoryKey() == "" {
		return errors.New("repository_key is required")
	}
	if o.GetPackageKey() == "" {
		return errors.New("package_key is required")
	}
	if o.GetAptPackage() == nil {
		return errors.New("apt_package is required")
	}
	if err := o.GetAptPackage().Validate(); err != nil {
		return err
	}
	if state := o.GetAptPackage().GetState(); state != AptPackageState_AptPackageState_IMPORTING {
		return errors.Wrap(ErrInvalidAptPackageInitialState, state.String())
	}
	return nil
}

// ApplyWorldOp applies the operation as a world operation.
func (o *AddAptPackageOp) ApplyWorldOp(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	sender peer.ID,
) (sysErr bool, err error) {
	// Validate the child record and its repository before changing the World.
	if err := o.Validate(); err != nil {
		return false, err
	}
	if err := world_types.CheckObjectType(ctx, ws, o.GetRepositoryKey(), AptRepositoryTypeID); err != nil {
		return false, err
	}

	// Create the package object from a copy of the requested record.
	packageKey := o.GetPackageKey()
	aptPackage := o.GetAptPackage().CloneVT()
	var createdObject world.ObjectState
	createdObject, _, err = world.CreateWorldObject(ctx, ws, packageKey, func(bcs *block.Cursor) error {
		bcs.SetBlock(aptPackage, true)
		return nil
	})
	world.ReleaseObjectState(createdObject)
	if err != nil {
		return false, err
	}

	// Register the package type and link it to its repository.
	if err := world_types.SetObjectType(ctx, ws, packageKey, AptPackageTypeID); err != nil {
		return false, err
	}
	if err := ws.SetGraphQuad(ctx, world.NewGraphQuadWithKeys(
		o.GetRepositoryKey(),
		PredAptRepoPackage.String(),
		packageKey,
		"",
	)); err != nil {
		return true, err
	}

	return false, nil
}

// NewAptPublishPackageOp constructs a new AptPublishPackageOp.
func NewAptPublishPackageOp(packageKey string) *AptPublishPackageOp {
	return &AptPublishPackageOp{PackageKey: packageKey}
}

// Validate performs cursory checks on the op.
func (o *AptPublishPackageOp) Validate() error {
	if o.GetPackageKey() == "" {
		return errors.New("package_key is required")
	}
	return nil
}

// ApplyWorldOp applies the operation as a world operation.
func (o *AptPublishPackageOp) ApplyWorldOp(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	sender peer.ID,
) (sysErr bool, err error) {
	if err := o.Validate(); err != nil {
		return false, err
	}
	err = applyAptPackageStateTransition(ctx, ws, o.GetPackageKey(), AptPackageState_AptPackageState_PUBLISHED)
	return false, err
}

// NewAptSupersedePackageOp constructs a new AptSupersedePackageOp.
func NewAptSupersedePackageOp(packageKey string) *AptSupersedePackageOp {
	return &AptSupersedePackageOp{PackageKey: packageKey}
}

// Validate performs cursory checks on the op.
func (o *AptSupersedePackageOp) Validate() error {
	if o.GetPackageKey() == "" {
		return errors.New("package_key is required")
	}
	return nil
}

// ApplyWorldOp applies the operation as a world operation.
func (o *AptSupersedePackageOp) ApplyWorldOp(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	sender peer.ID,
) (sysErr bool, err error) {
	if err := o.Validate(); err != nil {
		return false, err
	}
	err = applyAptPackageStateTransition(ctx, ws, o.GetPackageKey(), AptPackageState_AptPackageState_SUPERSEDED)
	return false, err
}

func applyAptPackageStateTransition(
	ctx context.Context,
	ws world.WorldState,
	packageKey string,
	next AptPackageState,
) error {
	// Open the typed package object for a state transition.
	if err := world_types.CheckObjectType(ctx, ws, packageKey, AptPackageTypeID); err != nil {
		return err
	}
	objectState, err := world.MustGetObject(ctx, ws, packageKey)
	defer world.ReleaseObjectState(objectState)
	if err != nil {
		return err
	}

	// Apply the package transition and persist its updated record.
	_, _, err = world.AccessObjectState(ctx, objectState, true, func(bcs *block.Cursor) error {
		// Decode the package record before changing its state.
		aptPackage, err := block.UnmarshalBlock[AptPackage](ctx, bcs)
		if err != nil {
			return err
		}

		// Validate the package transition and write the resulting record.
		if err := aptPackage.TransitionState(next); err != nil {
			return err
		}
		bcs.SetBlock(aptPackage, true)
		return nil
	})
	return err
}

// NewAddAptBuildSpecOp constructs a new AddAptBuildSpecOp.
func NewAddAptBuildSpecOp(repositoryKey, buildSpecKey string, buildSpec *AptBuildSpec) *AddAptBuildSpecOp {
	return &AddAptBuildSpecOp{
		RepositoryKey: repositoryKey,
		BuildSpecKey:  buildSpecKey,
		BuildSpec:     buildSpec,
	}
}

// Validate performs cursory checks on the op.
func (o *AddAptBuildSpecOp) Validate() error {
	// Require the repository and child record keys and contents.
	if o.GetRepositoryKey() == "" {
		return errors.New("repository_key is required")
	}
	if o.GetBuildSpecKey() == "" {
		return errors.New("build_spec_key is required")
	}
	if o.GetBuildSpec() == nil {
		return errors.New("build_spec is required")
	}
	return o.GetBuildSpec().Validate()
}

// ApplyWorldOp applies the operation as a world operation.
func (o *AddAptBuildSpecOp) ApplyWorldOp(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	sender peer.ID,
) (sysErr bool, err error) {
	// Validate the child record and its repository before changing the World.
	if err := o.Validate(); err != nil {
		return false, err
	}
	if err := world_types.CheckObjectType(ctx, ws, o.GetRepositoryKey(), AptRepositoryTypeID); err != nil {
		return false, err
	}

	// Create the build specification object from a copy of the request.
	buildSpecKey := o.GetBuildSpecKey()
	buildSpec := o.GetBuildSpec().CloneVT()
	var createdObject world.ObjectState
	createdObject, _, err = world.CreateWorldObject(ctx, ws, buildSpecKey, func(bcs *block.Cursor) error {
		bcs.SetBlock(buildSpec, true)
		return nil
	})
	world.ReleaseObjectState(createdObject)
	if err != nil {
		return false, err
	}

	// Register the build specification type and link it to its repository.
	if err := world_types.SetObjectType(ctx, ws, buildSpecKey, AptBuildSpecTypeID); err != nil {
		return false, err
	}
	if err := ws.SetGraphQuad(ctx, world.NewGraphQuadWithKeys(
		o.GetRepositoryKey(),
		PredAptRepoBuildSpec.String(),
		buildSpecKey,
		"",
	)); err != nil {
		return true, err
	}

	return false, nil
}

var (
	_ world.Operation = (*CreateAptRepositoryOp)(nil)
	_ world.Operation = (*AddAptPackageOp)(nil)
	_ world.Operation = (*AptPublishPackageOp)(nil)
	_ world.Operation = (*AptSupersedePackageOp)(nil)
	_ world.Operation = (*AddAptBuildSpecOp)(nil)
)
