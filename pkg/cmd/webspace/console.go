package webspace

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/containerd/console"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"

	"github.com/netsoc/cli/pkg/config"
	"github.com/netsoc/cli/pkg/util"
)

type consoleOptions struct {
	Config func() (*config.Config, error)

	User string
}

// NewCmdConsole creates a new webspace console command
func NewCmdConsole(f *util.CmdFactory) *cobra.Command {
	opts := consoleOptions{
		Config: f.Config,
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

	tty := console.Current()

	s, err := tty.Size()
	if err != nil {
		return fmt.Errorf("failed to get terminal size: %w", err)
	}
	if err := conn.WriteJSON(util.ConsoleSize{Width: int(s.Width), Height: int(s.Height)}); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send initial terminal size: %w", err)
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
		var ce *websocket.CloseError
		if errors.As(err, &ce) && ce.Code == websocket.CloseNormalClosure {
			fmt.Print("\r\n")
			return nil
		}

		return err
	}
}
