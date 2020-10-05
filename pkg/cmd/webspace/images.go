package webspace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type imagesOptions struct {
	WebspacedClient func() (*webspaced.APIClient, error)

	OutputFormat string
}

// NewCmdImages creates a new webspace images command
func NewCmdImages(f *util.CmdFactory) *cobra.Command {
	opts := imagesOptions{
		WebspacedClient: f.WebspacedClient,
	}
	cmd := &cobra.Command{
		Use:   "images",
		Short: "List available webspace images",
		RunE: func(cmd *cobra.Command, args []string) error {
			return imagesRun(opts)
		},
	}

	util.AddOptFormat(cmd, &opts.OutputFormat)

	return cmd
}

func printImages(images []webspaced.Image, outputType string) error {
	if strings.HasPrefix(outputType, "template=") {
		tpl, err := template.New("anonymous").Parse(strings.TrimPrefix(outputType, "template="))
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		if err := tpl.Execute(os.Stdout, images); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		return nil
	}

	switch outputType {
	case "json":
		if err := json.NewEncoder(os.Stdout).Encode(images); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
	case "yaml":
		if err := yaml.NewEncoder(os.Stdout).Encode(images); err != nil {
			return fmt.Errorf("failed to encode YAML: %w", err)
		}
	case "table", "table-wide", "wide":
		h := table.Row{"Alias", "Description", "Size"}
		if outputType == "table-wide" || outputType == "wide" {
			h = append(h, "Fingerprint")
		}

		t := table.NewWriter()
		t.AppendHeader(h)
		t.SetStyle(table.StyleRounded)

		for _, i := range images {
			alias := ""
			if len(i.Aliases) > 0 {
				alias = i.Aliases[0].Name
			}

			description, _ := i.Properties["description"]

			r := table.Row{alias, description, humanize.IBytes(uint64(i.Size))}
			if outputType == "table-wide" || outputType == "wide" {
				r = append(r, i.Fingerprint)
			}
			t.AppendRow(r)
		}

		fmt.Println(t.Render())
	default:
		return fmt.Errorf(`unknown output format "%v"`, outputType)
	}

	return nil
}

func imagesRun(opts imagesOptions) error {
	client, err := opts.WebspacedClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	images, _, err := client.ImagesApi.GetImages(ctx)
	if err != nil {
		return util.APIError(err)
	}

	return printImages(images, opts.OutputFormat)
}
