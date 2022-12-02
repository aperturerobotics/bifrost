package bifrost_rpc_access

import (
	"context"
	"regexp"

	bifrost_rpc "github.com/aperturerobotics/bifrost/rpc"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/refcount"
)

// ClientController resolves LookupRpcService with an AccessRpcService client.
type ClientController struct {
	info *controller.Info
	svc  BuildClientFunc

	clientRc *refcount.RefCount[*SRPCAccessRpcServiceClient]

	serviceIDRe *regexp.Regexp
	serverIDRe  *regexp.Regexp
}

// BuildClientFunc is a function to build the AccessRpcServiceClient.
// Returns the client, optional release function, and error.
type BuildClientFunc func(ctx context.Context) (*SRPCAccessRpcServiceClient, func(), error)

// NewBuildClientFunc constructs a BuildClientFunc with a static client.
func NewBuildClientFunc(svc SRPCAccessRpcServiceClient) BuildClientFunc {
	return func(ctx context.Context) (*SRPCAccessRpcServiceClient, func(), error) {
		return &svc, nil, nil
	}
}

// NewClientController constructs the controller.
// The regex fields can both be nil to accept any.
func NewClientController(
	info *controller.Info,
	svc BuildClientFunc,
	serviceIDRe *regexp.Regexp,
	serverIDRe *regexp.Regexp,
) *ClientController {
	c := &ClientController{
		info:        info,
		svc:         svc,
		serviceIDRe: serviceIDRe,
		serverIDRe:  serverIDRe,
	}
	c.clientRc = refcount.NewRefCount[*SRPCAccessRpcServiceClient](nil, nil, nil, svc)
	return c
}

// GetControllerInfo returns the controller info.
func (c *ClientController) GetControllerInfo() *controller.Info {
	return c.info.Clone()
}

// Execute executes the given controller.
func (c *ClientController) Execute(ctx context.Context) error {
	c.clientRc.SetContext(ctx)
	return nil
}

// HandleDirective asks if the handler can resolve the directive.
func (c *ClientController) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	switch dir := di.GetDirective().(type) {
	case bifrost_rpc.LookupRpcService:
		// filter by regex
		if c.serviceIDRe != nil {
			serviceID := dir.LookupRpcServiceID()
			if serviceID != "" && !c.serviceIDRe.MatchString(serviceID) {
				return nil, nil
			}
		}
		if c.serverIDRe != nil {
			serverID := dir.LookupRpcServerID()
			if serverID != "" && !c.serviceIDRe.MatchString(serverID) {
				return nil, nil
			}
		}
		return directive.R(NewLookupRpcServiceResolver(dir, c.BuildClient), nil)
	}
	return nil, nil
}

// BuildClient adds a reference to the client and waits for it to be built.
func (c *ClientController) BuildClient(ctx context.Context) (*SRPCAccessRpcServiceClient, func(), error) {
	client, ref, err := c.clientRc.Wait(ctx)
	if err != nil {
		return nil, nil, err
	}
	return client, ref.Release, nil
}

// Close releases any resources used by the controller.
func (c *ClientController) Close() error {
	return nil
}

// _ is a type assertion
var (
	_ controller.Controller = ((*ClientController)(nil))
	_ BuildClientFunc       = ((*ClientController)(nil).BuildClient)
)
