package webspace

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type initOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User       string
	Image      string
	NoPassword bool
	InstallSSH bool
}

// NewCmdInit creates a new webspace init command
func NewCmdInit(f *util.CmdFactory) *cobra.Command {
	opts := initOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}
	cmd := &cobra.Command{
		Use:     "init <image>",
		Aliases: []string{"create"},
		Short:   "Initialize webspace",
		Long: heredoc.Doc(`
			Initialize webspace using a provided image alias or fingerprint. By
			default sets root password by reading from stdin. Can also install
			an SSH server (providing an SSH key has been configured on the
			user's account), along with a port forward.
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Image = args[0]
			return initRun(opts)
		},
	}

	util.AddOptUser(cmd, &opts.User)
	cmd.Flags().BoolVar(&opts.NoPassword, "no-password", false, "don't set root password")
	cmd.Flags().BoolVarP(&opts.InstallSSH, "ssh", "s", false, "install SSH server")

	return cmd
}

func initRun(opts initOptions) error {
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

	var p string
	if !opts.NoPassword {
		p, err = util.ReadPassword(true)
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
	}

	req := webspaced.InitRequest{
		Image:    opts.Image,
		Password: p,
		Ssh:      opts.InstallSSH,
	}

	await, _, t := util.SimpleProgress("Initializing webspace", 10*time.Second)

	ws, _, err := client.ConfigApi.Create(ctx, opts.User, req)
	t.MarkAsDone()
	if err != nil {
		await()
		return util.APIError(err)
	}

	await()
	for e, i := range ws.Ports {
		if i == 22 {
			log.Printf("Webspace accessible over SSH on port %v", e)
		}
	}

	return nil
}
