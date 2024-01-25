package nats

import (
	nats_server "github.com/nats-io/nats-server/v2/server"
)

// clientAuth authenticates client connections.
type clientAuth struct{}

func newClientAuth() *clientAuth {
	return &clientAuth{}
}

// Check if a client is authorized to connect
func (a *clientAuth) Check(c nats_server.ClientAuthentication) bool {
	// TODO: always allow
	return true
}

// _ is a type assertion
var _ nats_server.Authentication = ((*clientAuth)(nil))
