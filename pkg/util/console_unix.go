// +build !windows

package util

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// ResizeListener listens for changes in console size
func ResizeListener(sizeChan chan ConsoleSize, stop chan struct{}) {
	winchChan := make(chan os.Signal, 1)
	signal.Notify(winchChan, syscall.SIGWINCH)

	stdin := int(os.Stdin.Fd())
	for {
		select {
		case <-winchChan:
			w, h, err := terminal.GetSize(stdin)
			if err != nil {
				log.Printf("Failed to get terminal size: %v", err)
				return
			}

			sizeChan <- ConsoleSize{w, h}
		case <-stop:
			return
		}
	}
}
