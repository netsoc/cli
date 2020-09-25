package account

import (
	"context"
	"fmt"
	"log"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	iam "github.com/netsoc/iam/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type loginOptions struct {
	Config    func() (*config.Config, error)
	IAMClient func() (*iam.APIClient, error)

	Force    bool
	Username string
}

// NewCmdLogin creates a new account login command
func NewCmdLogin(f *util.CmdFactory) *cobra.Command {
	opts := loginOptions{
		Config:    f.Config,
		IAMClient: f.IAMClient,
	}
	cmd := &cobra.Command{
		Use:   "login <username>",
		Short: "Log in to account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Username = args[0]
			return loginRun(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "log in even if already logged in")

	return cmd
}

func loginRun(opts loginOptions) error {
	c, err := opts.Config()
	if err != nil {
		return err
	}

	client, err := opts.IAMClient()
	if err != nil {
		return err
	}
	if c.Token != "" && !opts.Force {
		ctx := context.WithValue(context.Background(), iam.ContextAccessToken, c.Token)
		u, _, err := client.UsersApi.GetUser(ctx, "self")
		if err != nil {
			return fmt.Errorf("failed to get info about user: %w", util.APIError(err))
		}

		return fmt.Errorf(fmt.Sprintf("already logged in as %v", u.Username))
	}

	p, err := util.ReadPassword(false)
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	t, _, err := client.UsersApi.Login(context.Background(), opts.Username, iam.LoginRequest{Password: p})
	if err != nil {
		return util.APIError(err)
	}

	viper.Set("token", t.Token)
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	log.Println("Logged in successfully")

	return nil
}
