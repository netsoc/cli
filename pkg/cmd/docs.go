package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/MakeNowJust/heredoc"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const (
	sourceURL = "https://github.com/netsoc/cli"
	docsURL   = "https://docs.netsoc.ie/cli/"
)

// NewCmdDocs creates a new docs command
func NewCmdDocs() *cobra.Command {
	var docType, outDir string
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "View / generate documentation",
		Long: heredoc.Doc(`
			View or generate documentation for the Netsoc CLI. By default will
			open the latest online documentation on GitHub.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootCmd := cmd.Parent()

			switch docType {
			case "online":
				return open.Run(docsURL)
			case "man":
				outDir = path.Join(outDir, "man1")
				if err := os.MkdirAll(outDir, 0755); err != nil {
					return fmt.Errorf("failed to create output dir: %w", err)
				}

				return doc.GenManTree(rootCmd, &doc.GenManHeader{
					Title:   "NETSOC",
					Section: "1",
					Source:  sourceURL,
				}, outDir)
			case "markdown":
				return doc.GenMarkdownTree(rootCmd, outDir)
			case "rest":
				return doc.GenReSTTree(rootCmd, outDir)
			case "yaml":
				return doc.GenYamlTree(rootCmd, outDir)
			default:
				return fmt.Errorf("unsupported documentation type %v", docType)
			}
		},
	}

	cmd.Flags().StringVarP(&docType, "type", "t", "online", "documentation type `online|man|markdown|rest|yaml`")
	cmd.Flags().StringVarP(&outDir, "out-dir", "o", "docs", "generated documentation output directory")

	return cmd
}
