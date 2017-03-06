// +build windows

package console

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

const CommandExample = `
PS> openvdc console i-000000001
root@i-00000001#
`

func (s *SshConsole) getWinSize(fd uintptr) (int, int, error) {
	return getSize(fd)
}

func (s *SshConsole) signal(c chan<- os.Signal) error {
	signal.Notify(c, os.Interrupt)
	return nil
}

func (s *SshConsole) signalHandle(sig os.Signal, session *ssh.Session) error {
	switch sig {
	case os.Interrupt:
		sshSig := ssh.SIGINT
		if err := session.Signal(sshSig); err != nil {
			return errors.Wrapf(err, "Failed session.Signal: %s", sig)
		}
	default:
		return errors.Errorf("Unhandled signal: %s", sig)
	}
	return nil
}

// terminal.GetSize() on windows does not return the window dimension. It returns
// the console buffer depth can be seen in scrolled area.
// Stolen from golang.org/x/crypto/ssh/terminal/util_windows.go

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

var (
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

type (
	short int16
	word  uint16

	coord struct {
		x short
		y short
	}
	smallRect struct {
		left   short
		top    short
		right  short
		bottom short
	}
	consoleScreenBufferInfo struct {
		size              coord
		cursorPosition    coord
		attributes        word
		window            smallRect
		maximumWindowSize coord
	}
)

// GetSize returns the dimensions of the given terminal.
func getSize(fd uintptr) (width, height int, err error) {
	var info consoleScreenBufferInfo
	_, _, e := syscall.Syscall(procGetConsoleScreenBufferInfo.Addr(), 2, fd, uintptr(unsafe.Pointer(&info)), 0)
	if e != 0 {
		return 0, 0, error(e)
	}
	fmt.Println(info)
	return int(info.maximumWindowSize.x), int(info.maximumWindowSize.y), nil
}
