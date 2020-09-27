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
)

type deleteOptions struct {
	Config    func() (*config.Config, error)
	IAMClient func() (*iam.APIClient, error)

	NoConfirm bool
	User      string
}

// NewCmdDelete creates a new account delete command
func NewCmdDelete(f *util.CmdFactory) *cobra.Command {
	opts := deleteOptions{
		Config:    f.Config,
		IAMClient: f.IAMClient,
	}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete account",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.NoConfirm, "yes", false, "don't ask for confirmation")
	cmd.Flags().StringVarP(&opts.User, "user", "u", "self", "(admin only) user to delete")

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

	prompt := "Really delete yourself?"
	if opts.User != "self" {
		prompt = fmt.Sprintf("Really delete user %v?", opts.User)
	}

	if !opts.NoConfirm {
		shouldDelete, err := util.YesNo(prompt, false)
		if err != nil {
			return err
		}

		if !shouldDelete {
			return nil
		}
	}

	client, err := opts.IAMClient()
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), iam.ContextAccessToken, c.Token)

	if _, _, err := client.UsersApi.DeleteUser(ctx, opts.User); err != nil {
		return util.APIError(err)
	}

	log.Println("Deleted successfully")
	return nil
}
