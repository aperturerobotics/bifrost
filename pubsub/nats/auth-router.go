package nats

import (
	nats_server "github.com/nats-io/nats-server/v2/server"
)

// routerAuth authenticates client connections.
type routerAuth struct{}

func newRouterAuth() *routerAuth {
	return &routerAuth{}
}

// Check if a router is authorized to connect
func (a *routerAuth) Check(c nats_server.ClientAuthentication) bool {
	// TODO: always allow
	return true
}

// _ is a type assertion
var _ nats_server.Authentication = ((*routerAuth)(nil))
