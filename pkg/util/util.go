package util

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/go-github/v32/github"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/mattn/go-isatty"
	"github.com/netsoc/cli/version"
	iam "github.com/netsoc/iam/client"
	webspaced "github.com/netsoc/webspaced/client"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/yaml.v2"
)

var (
	// ErrPasswordMismatch indicates that a user entered two different passwords
	ErrPasswordMismatch = errors.New("passwords didn't match")
)

// IsDebug determines if debugging is enabled
var IsDebug bool

const (
	// TableDateFormat is the date format used for table output
	TableDateFormat = "2006-01-02 15:04:05"
	// DateOnlyFormat is a date format only containing the date
	DateOnlyFormat = "2006-01-02"
	// UpdateRepo is the repository to check for updates on
	UpdateRepo = "netsoc/cli"
)

// Debugf prints log messages only if debugging is enabled
func Debugf(format string, v ...interface{}) {
	if !IsDebug {
		return
	}

	log.Printf(format, v...)
}

// APIError re-formats an OpenAPI-generated API client error
func APIError(err error) error {
	var iamGeneric iam.GenericOpenAPIError
	if ok := errors.As(err, &iamGeneric); ok {
		if iamError, ok := iamGeneric.Model().(iam.Error); ok {
			return errors.New(iamError.Message)
		}
		return err
	}

	var wsdGeneric webspaced.GenericOpenAPIError
	if ok := errors.As(err, &wsdGeneric); ok {
		if wsdError, ok := wsdGeneric.Model().(webspaced.Error); ok {
			return errors.New(wsdError.Message)
		}
		return err
	}

	return err
}

// ReadPassword reads a password from stdin
func ReadPassword(confirm bool) (string, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		r := bufio.NewReader(os.Stdin)

		p := make([]byte, 1024)
		n, err := r.Read(p)
		if err != nil && !errors.Is(err, io.EOF) {
			return "", fmt.Errorf("read failed: %w", err)
		}

		return string(p[:n]), nil
	}

	fmt.Print("Enter password: ")
	p, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("read failed: %w", err)
	}
	fmt.Println()

	if confirm {
		fmt.Print("Again: ")
		p2, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", fmt.Errorf("read failed: %w", err)
		}
		fmt.Println()

		if string(p2) != string(p) {
			return "", ErrPasswordMismatch
		}
	}

	return string(p), nil
}

// YesNo asks a yes/no question on the command line
func YesNo(prompt string, yesDefault bool) (bool, error) {
	r := bufio.NewReader(os.Stdin)

	yn := "y/N"
	if yesDefault {
		yn = "Y/n"
	}
	for {
		fmt.Printf("%v [%v] ", prompt, yn)

		answer, err := r.ReadString('\n')
		if err != nil {
			return yesDefault, fmt.Errorf("read failed: %w", err)
		}

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "yes", "y":
			return true, nil
		case "no", "n":
			return false, nil
		case "":
			return yesDefault, nil
		}

		fmt.Println("Please enter y/n")
	}
}

// UserClaims represents claims in an auth JWT
type UserClaims struct {
	jwt.StandardClaims
	IsAdmin bool `json:"is_admin"`
	Version uint `json:"version"`
}

// PrintUsers renders a list of users (with various output options)
func PrintUsers(users []iam.User, outputType string, single bool) error {
	var data interface{}
	data = users
	if single && len(users) == 1 {
		data = users[0]
	}

	if strings.HasPrefix(outputType, "template=") {
		tpl, err := template.New("anonymous").Parse(strings.TrimPrefix(outputType, "template="))
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		if err := tpl.Execute(os.Stdout, data); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		return nil
	}

	switch outputType {
	case "json":
		if err := json.NewEncoder(os.Stdout).Encode(data); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
	case "yaml":
		if err := yaml.NewEncoder(os.Stdout).Encode(data); err != nil {
			return fmt.Errorf("failed to encode YAML: %w", err)
		}
	case "table-wide", "wide":
		t := table.NewWriter()
		t.AppendHeader(table.Row{"ID", "Username", "Admin", "Email", "Verified", "Name", "Renewed", "Created / Updated"})
		t.SetStyle(table.StyleRounded)
		t.Style().Options.SeparateRows = true

		for _, u := range users {
			admin := "no"
			if u.IsAdmin {
				admin = "yes"
			}

			verified := "no"
			if u.Verified {
				verified = "yes"
			}

			renewed := u.Renewed.Local().Format(TableDateFormat)
			if u.Renewed.Before(time.Unix(0, 0)) {
				renewed = "never"
			}

			createdUpdated := u.Meta.Created.Local().Format(TableDateFormat) + "\n" + u.Meta.Updated.Local().Format(TableDateFormat)

			t.AppendRow(table.Row{
				fmt.Sprint(u.Id),
				u.Username,
				admin,
				u.Email,
				verified,
				u.FirstName + " " + u.LastName,
				renewed,
				createdUpdated,
			})
		}

		fmt.Println(t.Render())
	case "table":
		t := table.NewWriter()
		t.AppendHeader(table.Row{"ID", "Username", "Email", "Name", "Renewed"})
		t.SetStyle(table.StyleRounded)

		for _, u := range users {
			renewed := u.Renewed.Local().Format(TableDateFormat)
			if u.Renewed.Before(time.Unix(0, 0)) {
				renewed = "never"
			}

			t.AppendRow(table.Row{fmt.Sprint(u.Id), u.Username, u.Email, u.FirstName + " " + u.LastName, renewed})
		}

		fmt.Println(t.Render())
	default:
		return fmt.Errorf(`unknown output format "%v"`, outputType)
	}

	return nil
}

// AddOptUser adds the user option to a command
func AddOptUser(cmd *cobra.Command, p *string) {
	cmd.Flags().StringVarP(p, "user", "u", "self", "(admin only) user to perform action as")
}

// AddOptFormat adds the output format option to a command
func AddOptFormat(cmd *cobra.Command, p *string) {
	cmd.Flags().StringVarP(p, "output", "o", "table", "output format `table|wide|yaml|json|template=<Go template>`")
}

// CheckUpdate checks to see if a new version is available
func CheckUpdate() (string, error) {
	current, err := semver.NewVersion(version.Version)
	if err != nil {
		// If the binary's version doesn't parse then it's probably a dev build
		return "", nil
	}

	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), "netsoc", "cli")
	if err != nil {
		return "", fmt.Errorf("failed to query GitHub API for latest release: %w", err)
	}

	new, err := semver.NewVersion(*release.TagName)
	if err != nil {
		return "", fmt.Errorf("failed to parse latest release tag: %w", err)
	}

	if new.GreaterThan(current) {
		return release.GetHTMLURL(), nil
	}

	return "", nil
}

// SimpleProgress renders a simple progress
func SimpleProgress(message string, eta time.Duration) (func(), progress.Writer, *progress.Tracker) {
	w := progress.NewWriter()
	w.SetAutoStop(true)
	w.ShowPercentage(false)
	w.ShowTime(true)
	w.ShowTracker(false)
	w.ShowValue(false)
	w.SetTrackerPosition(progress.PositionRight)
	go w.Render()

	t := &progress.Tracker{
		Message: message,
		Total:   1,
		Units:   progress.UnitsDefault,

		ExpectedDuration: eta,
	}
	w.AppendTracker(t)

	return func() {
		time.Sleep(250 * time.Millisecond)
	}, w, t
}
