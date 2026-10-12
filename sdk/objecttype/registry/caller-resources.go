package s4wave_objecttype_registry

import (
	"context"

	"github.com/aperturerobotics/starpc/srpc"
	resource_server "github.com/s4wave/spacewave/bldr/resource/server"
	"github.com/s4wave/spacewave/sdk/world/objecttype"
)

// callerResources preserves the authenticated caller through child publication.
type callerResources struct {
	// resources retains the underlying publication and release contract.
	resources resource_server.ResourceClientContext
	// caller is the granting dispatch identity.
	caller objecttype.Caller
}

// Context retains the caller through independently constructed descendants.
func (c *callerResources) Context() context.Context {
	return objecttype.WithCaller(c.resources.Context(), c.caller)
}

// AddResource publishes a child under the same authenticated caller.
func (c *callerResources) AddResource(invoker srpc.Invoker, release func()) (uint32, error) {
	return c.resources.AddResource(NewCallerInvoker(invoker, c.caller), release)
}

// AddResourceValue publishes a typed child under the same authenticated caller.
func (c *callerResources) AddResourceValue(invoker srpc.Invoker, value any, release func()) (uint32, error) {
	return c.resources.AddResourceValue(NewCallerInvoker(invoker, c.caller), value, release)
}

// ReleaseResource retains the underlying registration's cleanup contract.
func (c *callerResources) ReleaseResource(id uint32) bool {
	return c.resources.ReleaseResource(id)
}

// GetResourceValue returns the value retained by the underlying registration.
func (c *callerResources) GetResourceValue(id uint32) (any, error) {
	return c.resources.GetResourceValue(id)
}

// GetAttachedResource returns the client's attachment without changing custody.
func (c *callerResources) GetAttachedResource(id uint32) (srpc.Client, error) {
	return c.resources.GetAttachedResource(id)
}

// _ is a type assertion.
var _ resource_server.ResourceClientContext = (*callerResources)(nil)
