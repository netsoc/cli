package webspace

import (
	"context"
	"errors"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type syncOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User string
}

// NewCmdSync creates a new webspace sync command
func NewCmdSync(f *util.CmdFactory) *cobra.Command {
	opts := syncOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Re-generate webspace backend config",
		Long: heredoc.Doc(`
			Forces a reload of reverse proxy and port forwarding configuration.
			Useful if your username has been changed.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return syncRun(opts)
		},
	}

	util.AddOptUser(cmd, &opts.User)

	return cmd
}

func syncRun(opts syncOptions) error {
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

	_, err = client.StateApi.Sync(ctx, opts.User)
	if err != nil {
		return util.APIError(err)
	}

	return nil
}
