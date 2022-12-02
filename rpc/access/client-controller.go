package bifrost_rpc_access

import (
	"context"
	"regexp"

	bifrost_rpc "github.com/aperturerobotics/bifrost/rpc"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
)

// ClientController resolves LookupRpcService with an AccessRpcService client.
type ClientController struct {
	info        *controller.Info
	svc         SRPCAccessRpcServiceClient
	serviceIDRe *regexp.Regexp
	serverIDRe  *regexp.Regexp
}

// NewClientController constructs the controller.
// The regex fields can both be nil to accept any.
func NewClientController(
	info *controller.Info,
	svc SRPCAccessRpcServiceClient,
	serviceIDRe *regexp.Regexp,
	serverIDRe *regexp.Regexp,
) *ClientController {
	return &ClientController{
		info:        info,
		svc:         svc,
		serviceIDRe: serviceIDRe,
		serverIDRe:  serverIDRe,
	}
}

// GetControllerInfo returns the controller info.
func (c *ClientController) GetControllerInfo() *controller.Info {
	return c.info.Clone()
}

// Execute executes the given controller.
func (c *ClientController) Execute(ctx context.Context) error {
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
		return directive.R(NewLookupRpcServiceResolver(dir, c.svc), nil)
	}
	return nil, nil
}

// Close releases any resources used by the controller.
func (c *ClientController) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*ClientController)(nil))
