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

type portsOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	OutputFormat string
	User         string
}

// NewCmdPorts creates a new webspace ports command
func NewCmdPorts(f *util.CmdFactory) *cobra.Command {
	opts := portsOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}
	cmd := &cobra.Command{
		Use:   "ports",
		Short: "Configure webspace port forwards",
		RunE: func(cmd *cobra.Command, args []string) error {
			return portsRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "text", "output format `text|yaml|json|template=<Go template>`")
	util.AddOptUser(cmd, &opts.User)

	cmd.AddCommand(NewCmdPortsAdd(f), NewCmdPortsRemove(f))

	return cmd
}

func printPorts(ports map[string]int32, outputType string) error {
	if strings.HasPrefix(outputType, "template=") {
		tpl, err := template.New("anonymous").Parse(strings.TrimPrefix(outputType, "template="))
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		if err := tpl.Execute(os.Stdout, ports); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		return nil
	}

	switch outputType {
	case "json":
		if err := json.NewEncoder(os.Stdout).Encode(ports); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
	case "yaml":
		if err := yaml.NewEncoder(os.Stdout).Encode(ports); err != nil {
			return fmt.Errorf("failed to encode YAML: %w", err)
		}
	case "text":
		fmt.Println("Webspace port forwards:")
		for i, e := range ports {
			fmt.Printf(" - %v -> %v\n", i, e)
		}
	default:
		return fmt.Errorf(`unknown output format "%v"`, outputType)
	}

	return nil
}

func portsRun(opts portsOptions) error {
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

	ports, _, err := client.PortsApi.GetPorts(ctx, opts.User)
	if err != nil {
		return util.APIError(err)
	}

	return printPorts(ports, opts.OutputFormat)
}
