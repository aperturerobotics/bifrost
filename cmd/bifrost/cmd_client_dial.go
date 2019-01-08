package main

import (
	"context"
	"os"

	"github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/aperturerobotics/bifrost/stream/grpc/dial"
	"github.com/aperturerobotics/bifrost/stream/grpc/rpc"
	"github.com/aperturerobotics/bifrost/util/rwc"
	"github.com/urfave/cli"
)

var grpcdialConf stream_grpc_dial.Config

// runDialController runs a dial controller.
func runDialController(cctx *cli.Context) error {
	ctx := context.Background()
	c, err := GetClient()
	if err != nil {
		return err
	}

	client, err := c.DialStream(ctx)
	if err != nil {
		return err
	}

	err = client.Send(&stream_grpc.DialStreamRequest{
		Config: &grpcdialConf,
	})
	if err != nil {
		return err
	}

	rpc := stream_grpc.NewDialStreamClientRPC(client)
	return stream_grpc_rpc.AttachRPCToStream(
		rpc,
		rwc.NewReadWriteCloser(os.Stdin, os.Stdout),
		nil,
	)
}
