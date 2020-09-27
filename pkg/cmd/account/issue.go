package account

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	iam "github.com/netsoc/iam/client"
)

type issueOptions struct {
	Config    func() (*config.Config, error)
	IAMClient func() (*iam.APIClient, error)

	Username string
	Duration string
}

// NewCmdIssue creates a new account issue command
func NewCmdIssue(f *util.CmdFactory) *cobra.Command {
	opts := issueOptions{
		Config:    f.Config,
		IAMClient: f.IAMClient,
	}
	cmd := &cobra.Command{
		Use:   "issue <username> <duration>",
		Short: "(admin only) Issue a token",
		Long: heredoc.Doc(`
			Issue a token for a user. duration is a Go duration
			(see https://golang.org/pkg/time/#ParseDuration for details).
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Username = args[0]
			opts.Duration = args[1]

			return issueRun(opts)
		},
	}

	return cmd
}

func issueRun(opts issueOptions) error {
	c, err := opts.Config()
	if err != nil {
		return err
	}

	if c.Token == "" {
		return errors.New("not logged in")
	}

	client, err := opts.IAMClient()
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), iam.ContextAccessToken, c.Token)

	r, _, err := client.UsersApi.IssueToken(ctx, opts.Username, iam.IssueTokenRequest{Duration: opts.Duration})
	if err != nil {
		return util.APIError(err)
	}

	log.Println("New token:")
	fmt.Println(r.Token)

	return nil
}
