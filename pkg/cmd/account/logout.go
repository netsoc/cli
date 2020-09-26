package account

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	iam "github.com/netsoc/iam/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type logoutOptions struct {
	Config    func() (*config.Config, error)
	IAMClient func() (*iam.APIClient, error)

	All  bool
	User string
}

// NewCmdLogout creates a new account logout command
func NewCmdLogout(f *util.CmdFactory) *cobra.Command {
	opts := logoutOptions{
		Config:    f.Config,
		IAMClient: f.IAMClient,
	}
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out of account",
		RunE: func(cmd *cobra.Command, args []string) error {
			return logoutRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.All, "all", false, "log out of all devices")
	cmd.Flags().StringVarP(&opts.User, "user", "u", "self", "(admin only) user to log out, only applicable with --all")

	return cmd
}

func logoutRun(opts logoutOptions) error {
	c, err := opts.Config()
	if err != nil {
		return err
	}

	if c.Token == "" {
		return errors.New("not logged in")
	}

	if opts.All {
		client, err := opts.IAMClient()
		if err != nil {
			return err
		}
		ctx := context.WithValue(context.Background(), iam.ContextAccessToken, c.Token)

		u, _, err := client.UsersApi.GetUser(ctx, "self")
		if err != nil {
			return fmt.Errorf("failed to get info about user: %w", util.APIError(err))
		}

		_, err = client.UsersApi.Logout(ctx, opts.User)
		if err != nil {
			return util.APIError(err)
		}

		if opts.User != "self" && opts.User != u.Username {
			log.Println("Logged out successfully")
			return nil
		}
	} else if opts.User != "self" {
		return errors.New("user provided but `--all` not passed")
	}

	viper.Set("token", "")
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	log.Println("Logged out successfully")
	return nil
}
