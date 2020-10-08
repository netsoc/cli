package webspace

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
	webspaced "github.com/netsoc/webspaced/client"
)

type consoleOptions struct {
	Config          func() (*config.Config, error)
	WebspacedClient func() (*webspaced.APIClient, error)

	User string
}

// NewCmdConsole creates a new webspace console command
func NewCmdConsole(f *util.CmdFactory) *cobra.Command {
	opts := consoleOptions{
		Config:          f.Config,
		WebspacedClient: f.WebspacedClient,
	}
	cmd := &cobra.Command{
		Use:   "console",
		Short: "Attach to console",
		RunE: func(cmd *cobra.Command, args []string) error {
			return consoleRun(opts)
		},
	}

	util.AddOptUser(cmd, &opts.User)

	return cmd
}

func consoleRun(opts consoleOptions) error {
	c, err := opts.Config()
	if err != nil {
		return err
	}

	if c.Token == "" {
		return errors.New("not logged in")
	}

	log.Print("Attaching to console...")

	conn, err := util.WebspacedWebsocket(c, opts.User, "console")
	if err != nil {
		return fmt.Errorf("failed to open websocket connection: %w", err)
	}

	stdin := int(os.Stdin.Fd())

	w, h, err := util.GetTerminalSize(stdin)
	if err != nil {
		return fmt.Errorf("failed to get terminal size: %w", err)
	}
	if err := conn.WriteJSON(util.ConsoleSize{Width: w, Height: h}); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send initial terminal size: %w", err)
	}

	ttyState, err := terminal.MakeRaw(stdin)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to configure terminal: %w", err)
	}
	defer terminal.Restore(stdin, ttyState)

	rw := util.NewWebsocketIO(conn, func(s string, _ *util.WebsocketIO) {
		util.Debugf("Received websocket text message: %v", s)
	})
	defer rw.Close()

	errChan := make(chan error)
	resizeChan := make(chan util.ConsoleSize)
	stopResize := make(chan struct{})
	defer close(stopResize)
	go util.ResizeListener(resizeChan, stopResize)
	go func() {
		for {
			select {
			case s := <-resizeChan:
				rw.Mutex.Lock()
				util.Debugf("Sending console resize: %v", s)
				if err := conn.WriteJSON(s); err != nil {
					errChan <- err
				}
				rw.Mutex.Unlock()
			case <-stopResize:
				return
			}
		}
	}()

	escape := make(chan bool)
	er := util.NewEscapeReader(os.Stdin, escape)

	pipe := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errChan <- err
	}
	go pipe(rw, er)
	go pipe(os.Stdout, rw)

	log.Print("Attached, hit ^] (Ctrl+]) and then q to disconnect\r")

	select {
	case <-escape:
		fmt.Print("\r\n")
		return nil
	case err := <-errChan:
		return err
	}
}
