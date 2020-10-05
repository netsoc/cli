package cmd

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

// NewCmdCompletion creates a new completion command
func NewCmdCompletion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <bash|zsh|fish|powershell>",
		Short: "Generate shell completion scripts",
		Long: heredoc.Docf(`
			Generate shell completion scripts for Netsoc CLI commands.

			The output of this command will be a shell script and is meant to be saved to a
			file or immediately evaluated by an interactive shell.

			For example, for bash you could add this to your '~/.bash_profile':
				eval "$(%v completion bash)"

			When installing Netsoc CLI through a package manager, however, it's possible that
			no additional shell configuration is necessary to gain completion support.
		`, os.Args[0]),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shellType := args[0]
			rootCmd := cmd.Parent()

			switch shellType {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "powershell":
				return rootCmd.GenPowerShellCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			default:
				return fmt.Errorf("unsupported shell type %q", shellType)
			}
		},
	}

	return cmd
}
