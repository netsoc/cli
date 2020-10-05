package cmd

import (
	"fmt"
	"log"
	"path"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/netsoc/cli/pkg/cmd/account"
	"github.com/netsoc/cli/pkg/cmd/webspace"
	"github.com/netsoc/cli/pkg/util"
)

// NewCmdRoot creates a new root command
func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "netsoc",
		Short: "Netsoc CLI",
		Long:  `Manage Netsoc account and services from the command line.`,

		SilenceUsage: true,
	}

	defaultConfig := ".netsoc.yaml"
	home, err := homedir.Dir()
	if err == nil {
		defaultConfig = path.Join(home, defaultConfig)
	}

	cmd.PersistentFlags().String("config", defaultConfig, "config file")
	cmd.PersistentFlags().Bool("debug", false, "print debug messages")
	f := util.NewDefaultCmdFactory(cmd.PersistentFlags().Lookup("config"), cmd.PersistentFlags().Lookup("debug"))

	cmd.AddCommand(account.NewCmdAccount(f))
	cmd.AddCommand(webspace.NewCmdWebspace(f))
	cmd.AddCommand(NewCmdCompletion())
	cmd.AddCommand(NewCmdVersion(f))

	cmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		c, err := f.Config()
		if err != nil {
			return err
		}

		now := time.Now()
		if now.Sub(c.LastUpdateCheck).Hours() < 24 {
			return nil
		}

		newURL, err := util.CheckUpdate()
		if err != nil {
			return fmt.Errorf("Failed to check for updates: %v", err)
		}

		if newURL != "" {
			log.Printf("A new version of the Netsoc CLI is available at %v", newURL)
		}

		viper.Set("last_update_check", now)
		if err := viper.WriteConfig(); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		return nil
	}

	return cmd
}
