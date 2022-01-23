package cli

import (
	"encoding/json"
	"os"

	peer_grpc "github.com/aperturerobotics/bifrost/peer/rpc"
	"github.com/urfave/cli"
)

// RunPeerInfo runs the peer information command.
func (a *ClientArgs) RunPeerInfo(_ *cli.Context) error {
	ctx := a.GetContext()
	c, err := a.BuildClient()
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
