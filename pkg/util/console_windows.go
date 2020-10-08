// +build windows

package util

// ResizeListener on windows does nothing
func ResizeListener(sizeChan chan ConsoleSize, stop chan struct{}) {}
