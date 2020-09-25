package cmd

import (
	"path"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/cmd/account"
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
	cmd.AddCommand(NewCmdCompletion())
	cmd.AddCommand(NewCmdVersion(f))

	return cmd
}
