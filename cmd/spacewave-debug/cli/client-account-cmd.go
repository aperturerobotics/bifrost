//go:build !js

package cli

import (
	"context"
	"os"

	appcli "github.com/aperturerobotics/cli"
	"github.com/pkg/errors"

	resource "github.com/s4wave/spacewave/bldr/resource"
	resource_client "github.com/s4wave/spacewave/bldr/resource/client"
	s4wave_root "github.com/s4wave/spacewave/sdk/root"
	s4wave_session "github.com/s4wave/spacewave/sdk/session"
)

// AccountArgs contains account subcommand arguments.
type AccountArgs struct {
	client *ClientArgs

	// PluginID is the target plugin manifest ID.
	PluginID string
}

// BuildAccountCommand returns the account command with subcommands.
func (a *ClientArgs) BuildAccountCommand() *appcli.Command {
	aa := &AccountArgs{client: a}
	pluginFlag := func(usage string) []appcli.Flag {
		return []appcli.Flag{&appcli.StringFlag{
			Name:        "plugin-id",
			Usage:       usage,
			Required:    true,
			Destination: &aa.PluginID,
		}}
	}
	return &appcli.Command{
		Name:  "account",
		Usage: "manage the session account (plugins every Space loads)",
		Flags: a.BuildFlags(),
		Subcommands: []*appcli.Command{
			{
				Name:  "add-plugin",
				Usage: "add a plugin manifest ID to the account plugin list",
				Flags: pluginFlag("manifest ID of the plugin to add"),
				Action: func(c *appcli.Context) error {
					return aa.setPlugin(c.Context, true)
				},
			},
			{
				Name:  "remove-plugin",
				Usage: "remove a plugin manifest ID from the account plugin list",
				Flags: pluginFlag("manifest ID of the plugin to remove"),
				Action: func(c *appcli.Context) error {
					return aa.setPlugin(c.Context, false)
				},
			},
		},
	}
}

// setPlugin adds the plugin to or removes it from the account plugin list.
func (aa *AccountArgs) setPlugin(ctx context.Context, enabled bool) error {
	// Mount the selected Session.
	_, sess, release, err := aa.client.mountSession(ctx)
	if err != nil {
		return err
	}
	defer release()

	// Write the change and report it.
	if err := sess.SetAccountPlugin(ctx, aa.PluginID, enabled); err != nil {
		return errors.Wrap(err, "set account plugin")
	}
	verb := "removed plugin " + aa.PluginID + " from"
	if enabled {
		verb = "added plugin " + aa.PluginID + " to"
	}
	os.Stdout.WriteString(verb + " the account plugin list\n")
	return nil
}

// mountSession mounts the Session at the selected index through the core
// plugin's Resource service. Call release when done.
func (a *ClientArgs) mountSession(ctx context.Context) (*resource_client.Client, *s4wave_session.Session, func(), error) {
	// Connect to the core plugin that exposes the Resource service.
	coreClient, err := a.BuildCoreClient()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "connect to core plugin")
	}
	resClient, err := resource_client.NewClient(ctx, resource.NewSRPCResourceServiceClient(coreClient))
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "resource client")
	}

	// Access the root resource used to mount the selected Session.
	rootRef := resClient.AccessRootResource()
	root, err := s4wave_root.NewRoot(resClient, rootRef)
	if err != nil {
		rootRef.Release()
		resClient.Release()
		return nil, nil, nil, errors.Wrap(err, "root resource")
	}

	// Mount the selected Session.
	resp, err := root.MountSessionByIdx(ctx, sessionIndex32(a.SessionIdx))
	if err == nil && resp.GetNotFound() {
		err = errors.Errorf("no session at index %d", a.SessionIdx)
	}
	if err != nil {
		root.Release()
		resClient.Release()
		return nil, nil, nil, errors.Wrap(err, "mount session")
	}
	sessRef := resClient.CreateResourceReference(resp.GetResourceId())
	sess, err := s4wave_session.NewSession(resClient, sessRef)
	if err != nil {
		sessRef.Release()
		root.Release()
		resClient.Release()
		return nil, nil, nil, errors.Wrap(err, "session resource")
	}

	// Release the Session, root, and Resource client together.
	release := func() {
		sess.Release()
		root.Release()
		resClient.Release()
	}
	return resClient, sess, release, nil
}
