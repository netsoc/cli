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

type deleteOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	NoConfirm bool
	User      string
}

// NewCmdDelete creates a new webspace delete command
func NewCmdDelete(f *util.CmdFactory) *cobra.Command {
	opts := deleteOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}
	cmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"destroy"},
		Short:   "Delete webspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.NoConfirm, "yes", false, "don't ask for confirmation")
	util.AddOptUser(cmd, &opts.User)

	return cmd
}

func deleteRun(opts deleteOptions) error {
	c, err := opts.Config()
	if err != nil {
		return err
	}

	if c.Token == "" {
		return errors.New("not logged in")
	}

	if !opts.NoConfirm {
		shouldDelete, err := util.YesNo("Are you sure?", false)
		if err != nil {
			return err
		}

		if !shouldDelete {
			return nil
		}
	}

	client, err := opts.WebspacedClient()
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), webspaced.ContextAccessToken, c.Token)

	await, _, t := util.SimpleProgress("Deleting webspace", 5*time.Second)
	defer await()

	_, err = client.ConfigApi.Delete(ctx, opts.User)
	t.MarkAsDone()
	if err != nil {
		return util.APIError(err)
	}

	return nil
}
