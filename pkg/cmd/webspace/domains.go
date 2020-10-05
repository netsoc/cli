package webspace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type domainsOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	OutputFormat string
	User         string
}

// NewCmdDomains creates a new webspace domains command
func NewCmdDomains(f *util.CmdFactory) *cobra.Command {
	opts := domainsOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}
	cmd := &cobra.Command{
		Use:   "domains",
		Short: "Configure webspace domains",
		RunE: func(cmd *cobra.Command, args []string) error {
			return domainsRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "text", "output format `text|yaml|json|template=<Go template>`")
	util.AddOptUser(cmd, &opts.User)

	cmd.AddCommand(NewCmdDomainsAdd(f), NewCmdDomainsRemove(f))

	return cmd
}

func printDomains(domains []string, outputType string) error {
	if strings.HasPrefix(outputType, "template=") {
		tpl, err := template.New("anonymous").Parse(strings.TrimPrefix(outputType, "template="))
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		if err := tpl.Execute(os.Stdout, domains); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		return nil
	}

	switch outputType {
	case "json":
		if err := json.NewEncoder(os.Stdout).Encode(domains); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
	case "yaml":
		if err := yaml.NewEncoder(os.Stdout).Encode(domains); err != nil {
			return fmt.Errorf("failed to encode YAML: %w", err)
		}
	case "text":
		// There will always be at least the default domain
		fmt.Println("Webspace domains:")
		for _, d := range domains {
			fmt.Printf(" - %v\n", d)
		}
	default:
		return fmt.Errorf(`unknown output format "%v"`, outputType)
	}

	return nil
}

func domainsRun(opts domainsOptions) error {
	c, err := opts.Config()
	if err != nil {
		return err
	}

	if c.Token == "" {
		return errors.New("not logged in")
	}

	client, err := opts.WebspacedClient()
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), webspaced.ContextAccessToken, c.Token)

	domains, _, err := client.DomainsApi.GetDomains(ctx, opts.User)
	if err != nil {
		return util.APIError(err)
	}

	return printDomains(domains, opts.OutputFormat)
}
