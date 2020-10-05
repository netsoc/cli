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
	cmd.AddCommand(NewCmdInit(f), NewCmdDelete(f))
	// state
	cmd.AddCommand(NewCmdStatus(f), NewCmdStart(f), NewCmdReboot(f), NewCmdStop(f))
	// console
	cmd.AddCommand(NewCmdLog(f))

	return cmd
}
