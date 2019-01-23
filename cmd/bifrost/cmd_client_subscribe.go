package main

import (
	"context"
	"encoding/base64"
	"io"
	"os"
	"strconv"

	"github.com/aperturerobotics/bifrost/pubsub/grpc"
	"github.com/urfave/cli"
)

var subscribeConf pubsub_grpc.SubscribeRequest

// runSubscribe runs a pubsub subscription.
func runSubscribe(cctx *cli.Context) error {
	ctx := context.Background()
	c, err := GetClient()
	if err != nil {
		return err
	}

	client, err := c.Subscribe(ctx)
	if err != nil {
		return err
	}
	if err := client.Send(&subscribeConf); err != nil {
		return err
	}
	for {
		msg, err := client.Recv()
		if err != nil {
			if err != context.Canceled && err != io.EOF {
				os.Stderr.WriteString(err.Error())
				os.Stderr.WriteString("\n")
			}
			return err
		}
		if msg.GetSubscriptionStatus().GetSubscribed() {
			os.Stdout.WriteString("sub ")
			os.Stdout.WriteString(subscribeConf.GetChannelId())
			os.Stdout.WriteString("\n")
		}
		if msg.GetOutgoingStatus().GetIdentifier() != 0 {
			os.Stdout.WriteString("out")
			os.Stdout.WriteString(" ")
			os.Stdout.WriteString(
				strconv.Itoa(int(msg.GetOutgoingStatus().GetIdentifier())),
			)
			os.Stdout.WriteString(" ")
			mmsg := "ok"
			if !msg.GetOutgoingStatus().GetSent() {
				mmsg = "nok"
			}
			os.Stdout.WriteString(
				mmsg,
			)
			os.Stdout.WriteString("\n")
		}
		if len(msg.GetIncomingMessage().GetData()) != 0 {
			os.Stdout.WriteString(msg.GetIncomingMessage().GetFromPeerId())
			os.Stdout.WriteString(" ")
			authedMsg := "authenticated"
			if !msg.GetIncomingMessage().GetAuthenticated() {
				authedMsg = "unverified"
			}
			os.Stdout.WriteString(authedMsg)
			os.Stdout.WriteString(" ")
			os.Stdout.WriteString(
				base64.StdEncoding.EncodeToString(
					msg.GetIncomingMessage().GetData(),
				),
			)
			os.Stdout.WriteString("\n")
		}
	}
}
