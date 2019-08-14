package cli

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/aperturerobotics/bifrost/keypem"
	"github.com/aperturerobotics/bifrost/peer/grpc"
	"github.com/urfave/cli"
)

// RunIdentifyController runs an identify controller.
func (a *ClientArgs) RunIdentifyController(ctx context.Context, _ *cli.Context) error {
	c, err := a.BuildClient()
	if err != nil {
		return err
	}

	var dat []byte
	if a.IdentifyGenKey {
		if _, err := os.Stat(a.IdentifyKeyPath); os.IsNotExist(err) {
			privKey, _, err := keypem.GeneratePrivKey()
			if err != nil {
				return err
			}
			dat, err = keypem.MarshalPrivKeyPem(privKey)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(a.IdentifyKeyPath, dat, 0600); err != nil {
				return err
			}
		}
	}

	if len(dat) == 0 {
		dat, err = ioutil.ReadFile(a.IdentifyKeyPath)
		if err != nil {
			return err
		}
	}

	a.IdentifyConf.PrivKey = string(dat)
	if err := a.IdentifyConf.Validate(); err != nil {
		return err
	}

	req, err := c.Identify(ctx, &peer_grpc.IdentifyRequest{
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
