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

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type statusOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	OutputFormat string
	User         string
}

// NewCmdStatus creates a new webspace status command
func NewCmdStatus(f *util.CmdFactory) *cobra.Command {
	opts := statusOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"state"},
		Short:   "Get status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "text", "output format `text|yaml|json|template=<Go template>`")
	util.AddOptUser(cmd, &opts.User)

	return cmd
}

func printState(state webspaced.State, outputType string) error {
	if strings.HasPrefix(outputType, "template=") {
		tpl, err := template.New("anonymous").Parse(strings.TrimPrefix(outputType, "template="))
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		if err := tpl.Execute(os.Stdout, state); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		return nil
	}

	switch outputType {
	case "json":
		if err := json.NewEncoder(os.Stdout).Encode(state); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
	case "yaml":
		if err := yaml.NewEncoder(os.Stdout).Encode(state); err != nil {
			return fmt.Errorf("failed to encode YAML: %w", err)
		}
	case "text":
		running := "no"
		if state.Running {
			running = "yes"
		}
		fmt.Printf("Running: %v\n", running)

		if len(state.Usage.Disks) > 0 {
			fmt.Println("Disks:")
			for n, usage := range state.Usage.Disks {
				fmt.Printf(" - %v: Used %v\n", n, humanize.IBytes(uint64(usage)))
			}
		}

		if state.Running {
			fmt.Printf("CPU time: %v\n", time.Duration(state.Usage.Cpu))
			fmt.Printf("Memory usage: %v\n", humanize.IBytes(uint64(state.Usage.Memory)))
			fmt.Printf("Running processes: %v\n", state.Usage.Processes)

			if len(state.NetworkInterfaces) > 0 {
				fmt.Println("Network interfaces:")
				for n, iface := range state.NetworkInterfaces {
					fmt.Printf(" - %v (%v)\n", n, iface.Mac)
					fmt.Printf("   Sent/received: %v/%v\n",
						humanize.IBytes(uint64(iface.Counters.BytesSent)),
						humanize.IBytes(uint64(iface.Counters.BytesReceived)),
					)

					for _, addr := range iface.Addresses {
						t := "IPv4"
						if addr.Family == "inet6" {
							t = "IPv6"
						}

						fmt.Printf("   %v address: %v/%v\n", t, addr.Address, addr.Netmask)
					}
				}
			}
		}
	default:
		return fmt.Errorf(`unknown output format "%v"`, outputType)
	}

	return nil

}

func statusRun(opts statusOptions) error {
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

	state, _, err := client.StateApi.GetState(ctx, opts.User)
	if err != nil {
		return util.APIError(err)
	}

	return printState(state, opts.OutputFormat)
}
