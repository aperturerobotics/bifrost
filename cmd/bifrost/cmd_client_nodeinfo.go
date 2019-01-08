package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aperturerobotics/bifrost/peer/grpc"
	"github.com/aperturerobotics/controllerbus/grpc"
	"github.com/urfave/cli"
)

// clientCommands set in cmd_client.go

// runPeerInfo runs the peer information command.
func runPeerInfo(*cli.Context) error {
	ctx := context.Background()
	c, err := GetClient()
	if err != nil {
		return err
	}

	ni, err := c.GetPeerInfo(ctx, &peer_grpc.GetPeerInfoRequest{})
	if err != nil {
		return err
	}

	dat, err := json.MarshalIndent(ni, "", "\t")
	if err != nil {
		return err
	}

	os.Stdout.WriteString(string(dat))
	os.Stdout.WriteString("\n")
	return nil
}

// runBusInfo runs the bus information command.
func runBusInfo(*cli.Context) error {
	ctx := context.Background()
	c, err := GetClient()
	if err != nil {
		return err
	}

	ni, err := c.GetBusInfo(ctx, &controllerbus_grpc.GetBusInfoRequest{})
	if err != nil {
		return err
	}

	dat, err := json.MarshalIndent(ni, "", "\t")
	if err != nil {
		return err
	}

	os.Stdout.WriteString(string(dat))
	os.Stdout.WriteString("\n")
	return nil
}
