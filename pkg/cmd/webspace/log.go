package webspace

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type logOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User string
}

// NewCmdLog creates a new webspace log command
func NewCmdLog(f *util.CmdFactory) *cobra.Command {
	opts := logOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Get console log",
		RunE: func(cmd *cobra.Command, args []string) error {
			return logRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.User, "user", "u", "self", "(admin only) user to get log for")

	return cmd
}

func logRun(opts logOptions) error {
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

	log, _, err := client.ConsoleApi.GetLog(ctx, opts.User)
	if err != nil {
		return util.APIError(err)
	}

	fmt.Print(log)
	return nil
}
