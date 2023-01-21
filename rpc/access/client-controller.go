package bifrost_rpc_access

import (
	"context"
	"regexp"

	bifrost_rpc "github.com/aperturerobotics/bifrost/rpc"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/promise"
	"github.com/aperturerobotics/util/refcount"
)

// ClientController resolves LookupRpcService with an AccessRpcService client.
type ClientController struct {
	info *controller.Info
	svc  AccessClientFunc

	clientRc *refcount.RefCount[*SRPCAccessRpcServiceClient]

	serviceIDRe *regexp.Regexp
	serverIDRe  *regexp.Regexp
}

// AccessClientFunc is a function to access the AccessRpcServiceClient.
// The client should be released after the function returns.
// If the client is no longer valid, cancel the context.
type AccessClientFunc func(
	ctx context.Context,
	cb func(ctx context.Context, client SRPCAccessRpcServiceClient) error,
) error

// NewAccessClientFunc constructs a AccessClientFunc with a static client.
func NewAccessClientFunc(svc SRPCAccessRpcServiceClient) AccessClientFunc {
	return func(ctx context.Context, cb func(ctx context.Context, client SRPCAccessRpcServiceClient) error) error {
		return cb(ctx, svc)
	}
}

// NewClientController constructs the controller.
// The regex fields can both be nil to accept any.
func NewClientController(
	info *controller.Info,
	svc AccessClientFunc,
	serviceIDRe *regexp.Regexp,
	serverIDRe *regexp.Regexp,
) *ClientController {
	c := &ClientController{
		info:        info,
		svc:         svc,
		serviceIDRe: serviceIDRe,
		serverIDRe:  serverIDRe,
	}
	c.clientRc = refcount.NewRefCount(
		nil, nil, nil,
		func(ctx context.Context, released func()) (*SRPCAccessRpcServiceClient, func(), error) {
			clientCtx, clientCtxCancel := context.WithCancel(ctx)
			value := promise.NewPromise[*SRPCAccessRpcServiceClient]()
			go func() {
				err := svc(clientCtx, func(ctx context.Context, client SRPCAccessRpcServiceClient) error {
					value.SetResult(&client, nil)
					<-ctx.Done()
					released()
					return context.Canceled
				})
				if err != nil {
					value.SetResult(nil, err)
				}
			}()
			client, err := value.Await(ctx)
			if err != nil {
				clientCtxCancel()
				return nil, nil, err
			}
			return client, clientCtxCancel, nil
		},
	)
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
			if serverID != "" && !c.serverIDRe.MatchString(serverID) {
				return nil, nil
			}
		}
		return directive.R(NewLookupRpcServiceResolver(dir, c.AccessClient), nil)
	}
	return nil, nil
}

// AccessClient adds a reference to the client and waits for it to be built.
// Releases the client when the function returns.
func (c *ClientController) AccessClient(
	ctx context.Context,
	cb func(ctx context.Context, client SRPCAccessRpcServiceClient) error,
) error {
	return c.clientRc.Access(ctx, func(ctx context.Context, val *SRPCAccessRpcServiceClient) error {
		return cb(ctx, *val)
	})
}

// Close releases any resources used by the controller.
func (c *ClientController) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*ClientController)(nil))
var _ AccessClientFunc = ((*ClientController)(nil)).AccessClient
