package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/bifrost/keypem"
	nctr "github.com/aperturerobotics/bifrost/peer/controller"
	"github.com/urfave/cli"
)

var identifyConf nctr.Config
var identifyKeyPath string
var identifyGenKey bool

// runIdentifyController runs a identify controller.
func runIdentifyController(cctx *cli.Context) error {
	ctx := context.Background()

	var dat []byte
	if identifyGenKey {
		if _, err := os.Stat(identifyKeyPath); os.IsNotExist(err) {
			privKey, _, err := keypem.GeneratePrivKey()
			if err != nil {
				return err
			}
			dat, err = keypem.MarshalPrivKeyPem(privKey)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(identifyKeyPath, dat, 0600); err != nil {
				return err
			}
		}
	}

	var err error
	if len(dat) == 0 {
		dat, err = ioutil.ReadFile(identifyKeyPath)
		if err != nil {
			return err
		}
	}

	identifyConf.PrivKey = string(dat)
	if err := identifyConf.Validate(); err != nil {
		return err
	}

	c, err := GetClient()
	if err != nil {
		return err
	}

	req, err := c.Identify(ctx, &api.IdentifyRequest{
		Config: &identifyConf,
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
