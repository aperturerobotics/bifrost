//+build !js

package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/urfave/cli"
)

// clientCommands set in cmd_client.go

// runClientInfo runs the client information command.
func runClientInfo(*cli.Context) error {
	ctx := context.Background()
	c, err := GetClient()
	if err != nil {
		return err
	}

	ni, err := c.GetNodeInfo(ctx, &api.GetNodeInfoRequest{})
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
