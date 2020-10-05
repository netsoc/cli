package webspace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type configOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	OutputFormat string
	User         string
}

// NewCmdConfig creates a new webspace config command
func NewCmdConfig(f *util.CmdFactory) *cobra.Command {
	opts := configOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure webspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return configRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "text", "output format `text|yaml|json|template=<Go template>`")
	util.AddOptUser(cmd, &opts.User)

	cmd.AddCommand(NewCmdConfigSet(f))

	return cmd
}

func printConfig(config webspaced.Config, outputType string) error {
	if strings.HasPrefix(outputType, "template=") {
		tpl, err := template.New("anonymous").Parse(strings.TrimPrefix(outputType, "template="))
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		if err := tpl.Execute(os.Stdout, config); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		return nil
	}

	switch outputType {
	case "json":
		if err := json.NewEncoder(os.Stdout).Encode(config); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
	case "yaml":
		if err := yaml.NewEncoder(os.Stdout).Encode(config); err != nil {
			return fmt.Errorf("failed to encode YAML: %w", err)
		}
	case "text":
		fmt.Println("Webspace configuration:")
		fmt.Printf("Startup delay: %v\n", time.Duration(config.StartupDelay*1000*1000*1000))

		t := "HTTP"
		e := "disabled"
		if config.SniPassthrough {
			t = "HTTPS"
			e = "enabled"
		}
		fmt.Printf("%v port: %v\n", t, config.HttpPort)
		fmt.Printf("SNI passthrough is %v\n", e)
	default:
		return fmt.Errorf(`unknown output format "%v"`, outputType)
	}

	return nil

}

func configRun(opts configOptions) error {
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

	config, _, err := client.ConfigApi.GetConfig(ctx, opts.User)
	if err != nil {
		return util.APIError(err)
	}

	return printConfig(config, opts.OutputFormat)
}
