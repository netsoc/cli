// +build windows

package util

//import (
//	"os"
//
//	"github.com/TheTitanrain/w32"
//)

// ResizeListener on Windows does nothing
func ResizeListener(sizeChan chan ConsoleSize, stop chan struct{}) {}

// TODO: do this properly
// GetTerminalSize on Windows does nothing (Windows bad)
func GetTerminalSize(fd int) (int, int, error) {
	//stdout := w32.HANDLE(os.Stdout.Fd())
	//info := w32.GetConsoleScreenBufferInfo(stdout)

	//lines := info.SrWindow.Bottom - info.SrWindow.Top + 1
	//cols := info.SrWindow.Right - info.SrWindow.Left + 1
	//cols := info.DwSize.X
	//lines := info.DwSize.Y

	//return int(cols), int(lines), nil
	return 80, 24, nil
}
