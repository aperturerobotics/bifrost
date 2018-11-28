package main

import (
	"context"
	"os"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/aperturerobotics/bifrost/stream/grpc/accept"
	"github.com/aperturerobotics/bifrost/util/rwc"
	"github.com/urfave/cli"
)

var grpcacceptConf stream_grpc_accept.Config

// runAcceptController runs a accept controller.
func runAcceptController(cctx *cli.Context) error {
	ctx := context.Background()
	c, err := GetClient()
	if err != nil {
		return err
	}

	client, err := c.AcceptStream(ctx)
	if err != nil {
		return err
	}

	if len(remotePeerIdsCsv) != 0 {
		grpcacceptConf.RemotePeerIds = parseRemotePeerIdsCsv()
	}
	err = client.Send(&api.AcceptStreamRequest{
		Config: &grpcacceptConf,
	})
	if err != nil {
		return err
	}

	drpc := api.NewAcceptRPCClient(client)
	return stream_grpc.AttachRPCToStream(
		drpc,
		rwc.NewReadWriteCloser(os.Stdin, os.Stdout),
	)
}
