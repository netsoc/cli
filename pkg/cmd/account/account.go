package account

import (
	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/util"
)

// NewCmdAccount creates a new account management command
func NewCmdAccount(f *util.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account",
		Aliases: []string{"iam"},
		Short:   "Manage Netsoc account",
	}

	cmd.AddCommand(NewCmdLogin(f), NewCmdLogout(f), NewCmdInfo(f), NewCmdList(f))
	return cmd
}
