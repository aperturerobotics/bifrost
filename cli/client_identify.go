package cli

import (
	"os"

	peer_grpc "github.com/aperturerobotics/bifrost/peer/grpc"
	"github.com/urfave/cli"
)

// RunIdentifyController runs an identify controller.
func (a *ClientArgs) RunIdentifyController(_ *cli.Context) error {
	c, err := a.BuildClient()
	if err != nil {
		return err
	}

	dat, _, err := a.LoadOrGenerateIdentifyKey()
	a.IdentifyConf.PrivKey = string(dat)
	if err := a.IdentifyConf.Validate(); err != nil {
		return err
	}

	req, err := c.Identify(a.GetContext(), &peer_grpc.IdentifyRequest{
		Config: &a.IdentifyConf,
	})
	if err != nil {
		return err
	}

	for {
		resp, err := req.Recv()
		if err != nil {
			return err
		}

		os.Stdout.WriteString(resp.GetControllerStatus().String())
		os.Stdout.WriteString("\n")
	}
}
