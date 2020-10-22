// +build windows

package util

import "os"

// SignalForwardingSet is the set of signals that should can be forwarded
var SignalForwardingSet = []os.Signal{os.Interrupt}

// ResizeListener on Windows does nothing
func ResizeListener(sizeChan chan ConsoleSize, stop chan struct{}) {}

// GetTERM returns a hardcoded value on windows
func GetTERM() string {
	return "xterm"
}
