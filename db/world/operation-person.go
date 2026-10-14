package world

import "context"

// operationPersonKey keys the person an authenticated operation runs as.
type operationPersonKey struct{}

// operationOwnerKey keys whether that person owns the shared object.
type operationOwnerKey struct{}

// WithOperationPerson binds the accepted config's person to authenticated replay.
func WithOperationPerson(ctx context.Context, person string) context.Context {
	return context.WithValue(ctx, operationPersonKey{}, person)
}

// OperationPersonFromContext returns the person bound by the replay authority.
func OperationPersonFromContext(ctx context.Context) string {
	person, _ := ctx.Value(operationPersonKey{}).(string)
	return person
}

// WithOperationOwner records whether the accepted config makes the operation's
// signer an owner of the shared object, who may moderate content others wrote.
func WithOperationOwner(ctx context.Context, owner bool) context.Context {
	return context.WithValue(ctx, operationOwnerKey{}, owner)
}

// OperationOwnerFromContext reports whether the replay authority found the
// operation's signer to be an owner. It is false when none was recorded.
func OperationOwnerFromContext(ctx context.Context) bool {
	owner, _ := ctx.Value(operationOwnerKey{}).(bool)
	return owner
}
