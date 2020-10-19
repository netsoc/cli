package webspace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"text/template"

	"github.com/MakeNowJust/heredoc"
	"github.com/containerd/console"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type execOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User         string
	OutputFormat string
	Request      webspaced.ExecInteractiveRequest
}

// NewCmdExec creates a new webspace exec command
func NewCmdExec(f *util.CmdFactory) *cobra.Command {
	opts := execOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}

	env := []string{}
	cmd := &cobra.Command{
		Use:   "exec -- command [arg...]",
		Short: "Execute command in webspace",
		Long: heredoc.Doc(`
			Execute a command inside a webspace. By default runs interactively
			(with a PTY). Signals will be forwarded (SIGINT, SIGTERM etc.).

			If this command does not run in a TTY, the remote command will run
			non-interactively and the output will be YAML containing the
			captured stdout, stderr and exit code. (Use --output to use a
			different format)

			--uid, --gid, --env and --cwd only apply when running interactively.
		`),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Request.Command = args

			if !util.IsInteractive() && opts.OutputFormat == "interactive" {
				opts.OutputFormat = "yaml"
			}
			if opts.OutputFormat != "interactive" {
				if opts.Request.User != 0 || opts.Request.Group != 0 || len(env) != 0 || opts.Request.WorkingDirectory != "" {
					return fmt.Errorf("uid, gid, env and cwd only apply to interactive exec")
				}

				return execSimple(opts)
			}

			opts.Request.Environment = map[string]string{"TERM": util.GetTERM()}
			for _, e := range env {
				split := strings.Split(e, "=")
				if len(split) < 2 {
					return fmt.Errorf("invalid environment variable: %v", e)
				}

				opts.Request.Environment[split[0]] = strings.Join(split[1:], "=")
			}

			return execInteractive(opts)
		},
	}

	util.AddOptUser(cmd, &opts.User)
	cmd.Flags().StringVarP(&opts.OutputFormat, "output", "o", "interactive", "output format `interactive|yaml|json|template=<Go template>`")
	cmd.Flags().Int32Var(&opts.Request.User, "uid", 0, "webspace Linux user ID to run as")
	cmd.Flags().Int32Var(&opts.Request.Group, "gid", 0, "webspace Linux group ID to run as")
	cmd.Flags().StringArrayVarP(&env, "env", "e", []string{}, "environment variables to pass to command")
	cmd.Flags().StringVar(&opts.Request.WorkingDirectory, "cwd", "", "webspace command working directory")

	return cmd
}

func printSimple(result webspaced.ExecResponse, outputType string) error {
	if strings.HasPrefix(outputType, "template=") {
		tpl, err := template.New("anonymous").Parse(strings.TrimPrefix(outputType, "template="))
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		if err := tpl.Execute(os.Stdout, result); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		return nil
	}

	switch outputType {
	case "json":
		if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
	case "yaml":
		if err := yaml.NewEncoder(os.Stdout).Encode(result); err != nil {
			return fmt.Errorf("failed to encode YAML: %w", err)
		}
	}

	return nil
}
func execSimple(opts execOptions) error {
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

	result, _, err := client.ConsoleApi.Exec(ctx, opts.User, webspaced.ExecRequest{
		Command: strings.Join(opts.Request.Command, " "),
	})
	if err != nil {
		return util.APIError(err)
	}

	return printSimple(result, opts.OutputFormat)
}

func execInteractive(opts execOptions) error {
	c, err := opts.Config()
	if err != nil {
		return err
	}

	if c.Token == "" {
		return errors.New("not logged in")
	}

	conn, err := util.WebspacedWebsocket(c, opts.User, "exec")
	if err != nil {
		return fmt.Errorf("failed to open websocket connection: %w", err)
	}

	tty := console.Current()

	s, err := tty.Size()
	if err != nil {
		return fmt.Errorf("failed to get terminal size: %w", err)
	}
	opts.Request.Width = int32(s.Width)
	opts.Request.Height = int32(s.Height)

	if err := conn.WriteJSON(opts.Request); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send exec request: %w", err)
	}

	if err := tty.SetRaw(); err != nil {
		conn.Close()
		return fmt.Errorf("failed to put terminal in raw mode: %w", err)
	}
	defer tty.Reset()

	rw := util.NewWebsocketIO(conn, func(s string, _ *util.WebsocketIO) {
		util.Debugf("Received websocket text message: %v", s)
	})
	defer rw.Close()

	errChan := make(chan error)
	resizeChan := make(chan util.ConsoleSize)
	signalChan := make(chan os.Signal)
	stopControl := make(chan struct{})

	defer close(stopControl)
	go util.ResizeListener(resizeChan, stopControl)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTSTP,
		syscall.SIGTTIN, syscall.SIGTTOU, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGCONT)
	go func() {
		for {
			select {
			case s := <-resizeChan:
				rw.Mutex.Lock()

				util.Debugf("Sending console resize: %v", s)
				if err := conn.WriteJSON(webspaced.ExecInteractiveControl{
					Resize: webspaced.ResizeRequest{
						Width:  int32(s.Width),
						Height: int32(s.Height),
					},
				}); err != nil {
					errChan <- err
				}

				rw.Mutex.Unlock()
			case sig := <-signalChan:
				rw.Mutex.Lock()

				util.Debugf("Forwarding signal: %v", sig)
				if err := conn.WriteJSON(webspaced.ExecInteractiveControl{
					Signal: int32(util.SignalValue(sig)),
				}); err != nil {
					errChan <- err
				}

				rw.Mutex.Unlock()
			case <-stopControl:
				return
			}
		}
	}()

	pipe := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errChan <- err
	}
	go pipe(rw, os.Stdin)
	go pipe(os.Stdout, rw)

	var ce *websocket.CloseError
	err = <-errChan
	if errors.As(err, &ce) && ce.Code == websocket.CloseNormalClosure {
		util.ExitCode, err = strconv.Atoi(ce.Text)
		if err != nil {
			return fmt.Errorf("failed to parse exit code: %w", err)
		}

		return nil
	}

	return err
}

type loginOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User string
}

// NewCmdLogin creates a new webspace login command
func NewCmdLogin(f *util.CmdFactory) *cobra.Command {
	opts := loginOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Get shell in webspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(opts)
		},
	}

	util.AddOptUser(cmd, &opts.User)

	return cmd
}

func runLogin(opts loginOptions) error {
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

	result, _, err := client.ConsoleApi.Exec(ctx, opts.User, webspaced.ExecRequest{
		Command: "getent passwd root",
	})
	if err != nil {
		return util.APIError(err)
	}

	shell := "/bin/sh"
	if result.ExitCode != 0 {
		log.Printf("`getent passwd root` returned non-zero exit-code %v, guessing shell to be /bin/sh", result.ExitCode)
	} else {
		split := strings.Split(strings.TrimSpace(result.Stdout), ":")
		if len(split) != 7 {
			log.Printf("failed to parse getent output, guessing shell to be /bin/sh")
		}

		shell = split[6]
	}

	return execInteractive(execOptions{
		Config:          opts.Config,
		WebspacedClient: opts.WebspacedClient,

		User:         opts.User,
		OutputFormat: "interactive",
		Request: webspaced.ExecInteractiveRequest{
			Command: []string{shell},
		},
	})
}
