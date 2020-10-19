// +build windows

package util

// ResizeListener on Windows does nothing
func ResizeListener(sizeChan chan ConsoleSize, stop chan struct{}) {}

// GetTERM returns a hardcoded value on windows
func GetTERM() string {
	return "dumb"
}
