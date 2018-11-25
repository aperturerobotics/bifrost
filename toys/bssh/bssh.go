package bssh

import (
	"github.com/urfave/cli"
)

var privKeyPath string

var connectPeer string

func main() {
	app := cli.NewApp()
	app.Name = "bssh"
	app.Usage = "bssh is a bifrost ssh client"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "priv-key",
			Usage:       "path to private key, will be generated if doesn't exist",
			Destination: &privKeyPath,
			Value:       "bssh_priv.pem",
		},
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:  "connect",
			Usage: "connect to a remote bssh listener",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "peer",
					Usage:       "remote peer to connect to",
					Destination: &connectPeer,
				},
				// TODO: local-peer
			},
		},
		cli.Command{
			Name:  "listen",
			Usage: "listen for incoming bssh connections",
			Flags: []cli.Flag{},
		},
	}
}
