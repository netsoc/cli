package webspace

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type startOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User string
}

// NewCmdStart creates a new webspace start command
func NewCmdStart(f *util.CmdFactory) *cobra.Command {
	opts := startOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}
	cmd := &cobra.Command{
		Use:     "start",
		Aliases: []string{"boot"},
		Short:   "Boot webspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startRun(opts)
		},
	}

	util.AddOptUser(cmd, &opts.User)

	return cmd
}

func startRun(opts startOptions) error {
	c, err := opts.Config()
	if err != nil {
		return err
	}

	if c.Token == "" {
		return errors.New("not logged in")
	}

	client, err := opts.WebspacedClient()
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), webspaced.ContextAccessToken, c.Token)

	await, _, t := util.SimpleProgress("Starting webspace", 5*time.Second)
	defer await()

	_, err = client.StateApi.Start(ctx, opts.User)
	t.MarkAsDone()
	if err != nil {
		return util.APIError(err)
	}

	return nil
}
