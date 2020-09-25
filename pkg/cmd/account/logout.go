package account

import (
	"errors"
	"fmt"
	"log"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type logoutOptions struct {
	Config func() (*config.Config, error)
}

// NewCmdLogout creates a new account logout command
func NewCmdLogout(f *util.CmdFactory) *cobra.Command {
	opts := logoutOptions{
		Config: f.Config,
	}
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out of account",
		RunE: func(cmd *cobra.Command, args []string) error {
			return logoutRun(opts)
		},
	}

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

	viper.Set("token", "")
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	log.Println("Logged out successfully")

	return nil
}
