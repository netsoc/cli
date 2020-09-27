package account

import (
	"context"
	"errors"

	"github.com/MakeNowJust/heredoc"
	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	iam "github.com/netsoc/iam/client"
	"github.com/spf13/cobra"
)

type listOptions struct {
	Config    func() (*config.Config, error)
	IAMClient func() (*iam.APIClient, error)

	OutputFormat string
}

// NewCmdList creates a new account list command
func NewCmdList(f *util.CmdFactory) *cobra.Command {
	opts := listOptions{
		Config:    f.Config,
		IAMClient: f.IAMClient,
	}
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"users"},
		Short:   "(admin only) List users",
		Long: heredoc.Doc(`
			Prints details about users.

			A number of output format options are available.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listRun(opts)
		},
	}

	util.AddOptFormat(cmd, &opts.OutputFormat)

	return cmd
}

func listRun(opts listOptions) error {
	c, err := opts.Config()
	if err != nil {
		return err
	}

	if c.Token == "" {
		return errors.New("not logged in")
	}

	client, err := opts.IAMClient()
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), iam.ContextAccessToken, c.Token)

	u, _, err := client.UsersApi.GetUsers(ctx)
	if err != nil {
		return util.APIError(err)
	}

	return util.PrintUsers(u, opts.OutputFormat, false)
}
