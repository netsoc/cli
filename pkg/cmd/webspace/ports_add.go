package webspace

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type portsAddOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User         string
	ExternalPort uint16
	InternalPort uint16
}

// NewCmdPortsAdd creates a new webspace ports add command
func NewCmdPortsAdd(f *util.CmdFactory) *cobra.Command {
	opts := portsAddOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}

	cmd := &cobra.Command{
		Use:   "add <internal port>",
		Short: "Add port forward",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := strconv.ParseUint(args[0], 10, 16)
			if err != nil {
				return fmt.Errorf("failed to parse internal port: %w", err)
			}
			opts.InternalPort = uint16(p)

			return portsAddRun(opts)
		},
	}

	util.AddOptUser(cmd, &opts.User)
	cmd.Flags().Uint16VarP(&opts.ExternalPort, "external-port", "p", 0, "external port (0 means random)")

	return cmd
}

func portsAddRun(opts portsAddOptions) error {
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

	if opts.ExternalPort == 0 {
		i, _, err := client.PortsApi.AddRandomPort(ctx, opts.User, int32(opts.InternalPort))
		if err != nil {
			return util.APIError(err)
		}

		opts.ExternalPort = uint16(i.EPort)
	} else {
		if _, err := client.PortsApi.AddPort(ctx, opts.User, int32(opts.ExternalPort), int32(opts.InternalPort)); err != nil {
			return util.APIError(err)
		}
	}

	if util.IsInteractive() {
		fmt.Printf("Port %v in webspace is now accessible externally via port %v\n", opts.InternalPort, opts.ExternalPort)
	} else {
		fmt.Println(opts.ExternalPort)
	}

	return nil
}
