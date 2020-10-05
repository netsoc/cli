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

type shutdownOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User string
}

// NewCmdStop creates a new webspace shutdown command
func NewCmdStop(f *util.CmdFactory) *cobra.Command {
	opts := shutdownOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}
	cmd := &cobra.Command{
		Use:     "shutdown",
		Aliases: []string{"stop"},
		Short:   "Shut down webspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return shutdownRun(opts)
		},
	}

	util.AddOptUser(cmd, &opts.User)

	return cmd
}

func shutdownRun(opts shutdownOptions) error {
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

	await, _, t := util.SimpleProgress("Shutting down webspace", 5*time.Second)
	defer await()

	_, err = client.StateApi.Shutdown(ctx, opts.User)
	t.MarkAsDone()
	if err != nil {
		return util.APIError(err)
	}

	return nil
}
