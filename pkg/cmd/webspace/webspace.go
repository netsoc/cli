package webspace

import (
	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/util"
)

// NewCmdWebspace creates a new webspace management command
func NewCmdWebspace(f *util.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "webspace",
		Aliases: []string{"ws"},
		Short:   "Manage webspace",
	}

	// images
	cmd.AddCommand(NewCmdImages(f))
	// config
	cmd.AddCommand(NewCmdInit(f), NewCmdDelete(f), NewCmdConfig(f))
	// state
	cmd.AddCommand(NewCmdStatus(f), NewCmdStart(f), NewCmdSync(f), NewCmdReboot(f), NewCmdStop(f))
	// domains
	cmd.AddCommand(NewCmdDomains(f))
	// ports
	cmd.AddCommand(NewCmdPorts(f))
	// console
	cmd.AddCommand(NewCmdLog(f), NewCmdConsole(f), NewCmdExec(f), NewCmdLogin(f))

	return cmd
}
