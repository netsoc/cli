package account

import (
	"context"
	"errors"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	iam "github.com/netsoc/iam/client"
)

type infoOptions struct {
	Config    func() (*config.Config, error)
	IAMClient func() (*iam.APIClient, error)

	OutputFormat string
	User         string
}

// NewCmdInfo creates a new account info command
func NewCmdInfo(f *util.CmdFactory) *cobra.Command {
	opts := infoOptions{
		Config:    f.Config,
		IAMClient: f.IAMClient,
	}
	cmd := &cobra.Command{
		Use:     "info",
		Aliases: []string{"get"},
		Short:   "Get info about user",
		Long: heredoc.Doc(`
			Prints details about the user. By default, retrieves the logged-in user's profile.
			If logged in with an admin account, the -u flag can be used to select the user to retrieve.

			A number of output format options are available.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return infoRun(opts)
		},
	}

	util.AddOptFormat(cmd, &opts.OutputFormat)
	cmd.Flags().StringVarP(&opts.User, "user", "u", "self", "(admin only) user to get info about")

	return cmd
}

func infoRun(opts infoOptions) error {
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

	u, _, err := client.UsersApi.GetUser(ctx, opts.User)
	if err != nil {
		return util.APIError(err)
	}

	return util.PrintUsers([]iam.User{u}, opts.OutputFormat, true)
}
