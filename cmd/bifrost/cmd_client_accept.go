package main

import (
	"context"
	"os"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/aperturerobotics/bifrost/stream/grpc/accept"
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

	errCh := make(chan error, 3)

	// write stdin -> request
	go func() {
		data := make([]byte, 1500)
		for {
			n, err := os.Stdin.Read(data)
			if err != nil {
				errCh <- err
				return
			}

			err = client.Send(&api.AcceptStreamRequest{
				Request: &stream_grpc.Request{
					Data: data[:n],
				},
			})
			if err != nil {
				errCh <- err
				return
			}
		}
	}()

	// write responses -> stdout
	for {
		resp, err := client.Recv()
		if err != nil {
			return err
		}

		os.Stdout.Write(resp.GetResponse().GetData())
	}
}
