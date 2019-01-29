package main

import (
	"net"
	"os"
	"strings"

	"github.com/aperturerobotics/bifrost/stream/grpc/accept"
	"github.com/aperturerobotics/bifrost/stream/grpc/dial"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
)

var privKeyPath string
var clientDialAddr string
var connectPeer string

var ProtocolID = "bssh/1"

var ExecProcess = "/bin/bash --login"

var grpcdialConf = &stream_grpc_dial.Config{
	Encrypted: true,
	Reliable:  true,
}

var grpcacceptConf = &stream_grpc_accept.Config{}

var sshServerConf = &ssh.ServerConfig{
	ServerVersion: ProtocolID,
	NoClientAuth:  true,
}

var sshClientConf = &ssh.ClientConfig{
	User: "testuser",
	HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	},
}

var remotePeerIdsCsv string

func parseRemotePeerIdsCsv() []string {
	pts := strings.Split(remotePeerIdsCsv, ",")
	var peerIds []string
	for _, pt := range pts {
		pt = strings.TrimSpace(pt)
		peerIds = append(peerIds, pt)
	}
	return peerIds
}

func main() {
	app := cli.NewApp()
	app.Name = "bssh"
	app.Usage = "bssh is a bifrost ssh implementation on top of the API"
	app.HideVersion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "priv-key",
			Usage:       "path to private key, will be generated if doesn't exist",
			Destination: &privKeyPath,
			Value:       "bssh_priv.pem",
		},
		cli.StringFlag{
			Name:        "dial-addr",
			Usage:       "address to dial API on",
			Destination: &clientDialAddr,
			Value:       "localhost:5110",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "connect",
			Usage:  "connect to a remote bssh listener",
			Action: runConnect,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "local-peer-id",
					Usage:       "local peer ID to dial from, can be empty",
					Destination: &grpcdialConf.LocalPeerId,
				},
				&cli.StringFlag{
					Name:        "protocol-id",
					Usage:       "protocol ID to use",
					Destination: &grpcdialConf.ProtocolId,
					Value:       ProtocolID,
				},
				&cli.StringFlag{
					Name:        "peer-id",
					Usage:       "remote peer id to dial",
					Destination: &grpcdialConf.PeerId,
				},
				&cli.Uint64Flag{
					Name:        "transport-id",
					Usage:       "if set, filter the transport id",
					Destination: &grpcdialConf.TransportId,
				},
			},
		},
		{
			Name:   "listen",
			Usage:  "listen for incoming bssh connections",
			Action: runListen,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "local-peer-id",
					Usage:       "local peer ID to match incoming streams to",
					Destination: &grpcacceptConf.LocalPeerId,
				},
				&cli.StringFlag{
					Name:        "process",
					Usage:       "process to execute",
					Destination: &ExecProcess,
					Value:       ExecProcess,
				},
				&cli.StringFlag{
					Name:        "protocol-id",
					Usage:       "protocol ID to use",
					Destination: &grpcacceptConf.ProtocolId,
					Value:       ProtocolID,
				},
				&cli.StringFlag{
					Name:        "remote-peer-ids",
					Usage:       "remote peer ids, comma separated, to allow to connect",
					Destination: &remotePeerIdsCsv,
				},
				&cli.Uint64Flag{
					Name:        "transport-id",
					Usage:       "if set, filter the transport id",
					Destination: &grpcacceptConf.TransportId,
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
}
