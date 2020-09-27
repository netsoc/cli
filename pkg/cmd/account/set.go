package account

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/MakeNowJust/heredoc"
	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	iam "github.com/netsoc/iam/client"
)

type setOptions struct {
	Config    func() (*config.Config, error)
	IAMClient func() (*iam.APIClient, error)

	User     string
	Property string
	Value    string
}

// NewCmdSet creates a new account set command
func NewCmdSet(f *util.CmdFactory) *cobra.Command {
	opts := setOptions{
		Config:    f.Config,
		IAMClient: f.IAMClient,
	}

	listStyle := list.StyleBulletCircle
	listStyle.LinePrefix = "  "

	settable := list.NewWriter()
	settable.SetStyle(listStyle)
	settable.AppendItems([]interface{}{"username", "email", "password", "firstname", "lastname"})

	admin := list.NewWriter()
	admin.SetStyle(listStyle)
	admin.AppendItems([]interface{}{"verified", "renewed", "isadmin"})

	causesRoll := list.NewWriter()
	causesRoll.SetStyle(listStyle)
	causesRoll.AppendItems([]interface{}{"password", "email", "isadmin"})

	cmd := &cobra.Command{
		Use:     "set <property> <value>",
		Aliases: []string{"get"},
		Short:   "Set user property",
		Args:    cobra.ExactArgs(2),
		Long: heredoc.Docf(`
			Set user property.
			If logged in with an admin account, the -u flag can be used to
			select the user to retrieve. Dates are of the form "YYYY-MM-DD".

			The following properties can be set:
			%v

			These properties can only be set by an admin:
			%v

			Setting the following will log out *all* existing sessions (setting
			email will also re-send verification):
			%v
		`, settable.Render(), admin.Render(), causesRoll.Render()),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Property = args[0]
			opts.Value = args[1]

			return setRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.User, "user", "u", "self", "(admin only) user to set property on")

	return cmd
}

func setRun(opts setOptions) error {
	c, err := opts.Config()
	if err != nil {
		return err
	}

	if c.Token == "" {
		return errors.New("not logged in")
	}

	patch := map[string]string{opts.Property: opts.Value}
	var patchUser iam.User
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused:      true,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(util.DateOnlyFormat),
		),

		Result: &patchUser,
	})
	if err != nil {
		return fmt.Errorf("failed to create payload encoder: %w", err)
	}
	if err := decoder.Decode(patch); err != nil {
		return fmt.Errorf("failed to create update payload: %w", err)
	}

	client, err := opts.IAMClient()
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), iam.ContextAccessToken, c.Token)

	if _, _, err := client.UsersApi.UpdateUser(ctx, opts.User, patchUser); err != nil {
		return util.APIError(err)
	}

	log.Print("Updated successfully")
	return nil
}
