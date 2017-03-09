// +build windows

package console

import (
	"io"
	"os"
	"os/signal"

	"github.com/Azure/go-ansiterm"
	"github.com/Azure/go-ansiterm/winterm"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

const CommandExample = `
PS> openvdc console i-000000001
root@i-00000001#
`

func (s *SshConsole) bindFDs(session *ssh.Session) error {
	session.Stdin = os.Stdin
	session.Stdout = newAnsiWriter(os.Stdout)
	session.Stderr = newAnsiWriter(os.Stderr)
	return nil
}

func (s *SshConsole) getWinSize(fd uintptr) (int, int, error) {
	info, err := winterm.GetConsoleScreenBufferInfo(fd)
	if err != nil {
		return 0, 0, errors.Wrap(err, "winterm.GetnConsoleScreenBufferInfo")
	}
	return int(info.Window.Right - info.Window.Left + 1), int(info.Window.Bottom - info.Window.Top + 1), nil
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

// Stolen from github.com/docker/docker/pkg/term/windows/ansi_writer.go

func newAnsiWriter(out *os.File) io.Writer {
	return &ansiWriter{
		parser: ansiterm.CreateParser("Ground", winterm.CreateWinEventHandler(out.Fd(), out)),
	}
}

type ansiWriter struct {
	parser *ansiterm.AnsiParser
}

func (w *ansiWriter) Write(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	return w.parser.Parse(buf)
}
