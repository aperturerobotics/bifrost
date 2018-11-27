package main

import (
	"context"
	"os"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/aperturerobotics/bifrost/stream/grpc/dial"
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

	err = client.Send(&api.DialStreamRequest{
		Config: &grpcdialConf,
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

			err = client.Send(&api.DialStreamRequest{
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
