// +build !windows

package util

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/containerd/console"
)

// ResizeListener listens for changes in console size
func ResizeListener(sizeChan chan ConsoleSize, stop chan struct{}) {
	winchChan := make(chan os.Signal, 1)
	signal.Notify(winchChan, syscall.SIGWINCH)

	tty := console.Current()
	for {
		select {
		case <-winchChan:
			s, err := tty.Size()
			if err != nil {
				log.Printf("Failed to get terminal size: %v", err)
				return
			}

			sizeChan <- ConsoleSize{int(s.Width), int(s.Height)}
		case <-stop:
			return
		}
	}
}

// GetTERM gets the value of the TERM environment variable
func GetTERM() string {
	t, _ := os.LookupEnv("TERM")
	return t
}
