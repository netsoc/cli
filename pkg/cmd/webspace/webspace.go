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

	cmd.AddCommand(NewCmdImages(f), NewCmdInit(f), NewCmdDelete(f))

	return cmd
}
