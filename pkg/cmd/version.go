package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/util"
	"github.com/netsoc/cli/version"
)

// NewCmdVersion creates a new version command
func NewCmdVersion(f *util.CmdFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return versionRun(f)
		},
	}
}

func versionRun(f *util.CmdFactory) error {
	if _, err := f.Config(); err != nil {
		return err
	}

	log.Printf("CLI version: %v", version.Version)
	log.Printf("IAM API version: %v", version.IAM)
	log.Printf("webspaced API version: %v", version.Webspaced)
	return nil
}
