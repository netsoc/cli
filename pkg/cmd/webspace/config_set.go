package webspace

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
	webspaced "github.com/netsoc/webspaced/client"
)

type configSetOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User   string
	Option string
	Value  string
}

// NewCmdConfigSet creates a new webspace config set command
func NewCmdConfigSet(f *util.CmdFactory) *cobra.Command {
	opts := configSetOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}

	listStyle := list.StyleBulletCircle
	listStyle.LinePrefix = "  "

	settable := list.NewWriter()
	settable.SetStyle(listStyle)
	settable.AppendItems([]interface{}{"startupDelay", "httpPort", "sniPassthrough"})

	cmd := &cobra.Command{
		Use:   "set <property> <value>",
		Short: "Set config option",
		Args:  cobra.ExactArgs(2),
		Long: heredoc.Docf(`
			Set config option.

			The following options can be set:
			%v

			httpPort will be used as the TLS port if SNI passthrough is enabled.
		`, settable.Render()),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Option = args[0]
			opts.Value = args[1]

			return configSetRun(opts)
		},
	}

	util.AddOptUser(cmd, &opts.User)

	return cmd
}

func configSetRun(opts configSetOptions) error {
	c, err := opts.Config()
	if err != nil {
		return err
	}

	if c.Token == "" {
		return errors.New("not logged in")
	}

	patch := map[string]string{opts.Option: opts.Value}
	var patchConfig webspaced.Config
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused:      true,
		WeaklyTypedInput: true,

		Result: &patchConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to create payload encoder: %w", err)
	}
	if err := decoder.Decode(patch); err != nil {
		return fmt.Errorf("failed to create update payload: %w", err)
	}

	client, err := opts.WebspacedClient()
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), webspaced.ContextAccessToken, c.Token)

	if _, _, err := client.ConfigApi.UpdateConfig(ctx, opts.User, patchConfig); err != nil {
		return util.APIError(err)
	}

	log.Print("Updated successfully")
	return nil
}
