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

type rebootOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User string
}

// NewCmdReboot creates a new webspace reboot command
func NewCmdReboot(f *util.CmdFactory) *cobra.Command {
	opts := rebootOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}
	cmd := &cobra.Command{
		Use:     "reboot",
		Aliases: []string{"restart"},
		Short:   "Reboot webspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return rebootRun(opts)
		},
	}

	util.AddOptUser(cmd, &opts.User)

	return cmd
}

func rebootRun(opts rebootOptions) error {
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

	await, _, t := util.SimpleProgress("Rebooting webspace", 5*time.Second)
	defer await()

	_, err = client.StateApi.Reboot(ctx, opts.User)
	t.MarkAsDone()
	if err != nil {
		return util.APIError(err)
	}

	return nil
}
