package webspace

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type portsRemoveOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User string
	Port uint16
}

// NewCmdPortsRemove creates a new webspace ports remove command
func NewCmdPortsRemove(f *util.CmdFactory) *cobra.Command {
	opts := portsRemoveOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}

	cmd := &cobra.Command{
		Use:     "remove <external port>",
		Aliases: []string{"delete"},
		Short:   "Remove port forward",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := strconv.ParseUint(args[0], 10, 16)
			if err != nil {
				return fmt.Errorf("failed to parse port: %w", err)
			}
			opts.Port = uint16(p)

			return portsRemoveRun(opts)
		},
	}

	util.AddOptUser(cmd, &opts.User)

	return cmd
}

func portsRemoveRun(opts portsRemoveOptions) error {
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

	if _, err := client.PortsApi.RemovePort(ctx, opts.User, int32(opts.Port)); err != nil {
		return util.APIError(err)
	}

	log.Print("Removed successfully")
	return nil
}
