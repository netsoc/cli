package webspace

import (
	"context"
	"errors"
	"log"

	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type domainsRemoveOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User   string
	Domain string
}

// NewCmdDomainsRemove creates a new webspace domains remove command
func NewCmdDomainsRemove(f *util.CmdFactory) *cobra.Command {
	opts := domainsRemoveOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}

	cmd := &cobra.Command{
		Use:     "remove <domain>",
		Aliases: []string{"delete"},
		Short:   "Add custom domain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Domain = args[0]
			return domainsRemoveRun(opts)
		},
	}

	util.AddOptUser(cmd, &opts.User)

	return cmd
}

func domainsRemoveRun(opts domainsRemoveOptions) error {
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

	if _, err := client.DomainsApi.RemoveDomain(ctx, opts.User, opts.Domain); err != nil {
		return util.APIError(err)
	}

	log.Print("Removed successfully")
	return nil
}
