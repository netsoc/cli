package webspace

import (
	"context"
	"errors"
	"log"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type domainsAddOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User   string
	Domain string
}

// NewCmdDomainsAdd creates a new webspace domains add command
func NewCmdDomainsAdd(f *util.CmdFactory) *cobra.Command {
	opts := domainsAddOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}

	cmd := &cobra.Command{
		Use:   "add <domain>",
		Short: "Add custom domain",
		Args:  cobra.ExactArgs(1),
		Long: heredoc.Doc(`
			Add custom domain.

			Domain will be verified by looking for a TXT record of the format
			webspace:id:<user id>.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Domain = args[0]
			return domainsAddRun(opts)
		},
	}

	util.AddOptUser(cmd, &opts.User)

	return cmd
}

func domainsAddRun(opts domainsAddOptions) error {
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

	if _, err := client.DomainsApi.AddDomain(ctx, opts.User, opts.Domain); err != nil {
		return util.APIError(err)
	}

	log.Print("Verified successfully")
	return nil
}
