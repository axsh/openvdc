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

func (s *SshConsole) bindFDs(session *ssh.Session) (func() error, error) {
	closeFunc := func() error { return nil }
	caps := detectConsoleApp()
	if modes, err := winterm.GetConsoleMode(os.Stdout.Fd()); err == nil && modes != 0 {
		// GetConsoleScreenBufferInfo() fails on STD_INPUT_HANDLE due to
		// missing GENERIC_READ access right. So we use os.Stdout.Fd().
		// https://msdn.microsoft.com/en-us/library/ms683171(VS.85).aspx
		origModes, err := makeRaw(os.Stdout.Fd())
		if err != nil {
			return nil, err
		}

		closeFunc = func() error {
			return restore(os.Stdout.Fd(), origModes)
		}
	}

	if caps.EmulateStdin {
		session.Stdin = newAnsiReader(os.Stdin)
	} else {
		session.Stdin = os.Stdin
	}
	if caps.EmulateStdout {
		session.Stdout = newAnsiWriter(os.Stdout)
	} else {
		session.Stdout = os.Stdout
	}
	if caps.EmulateStderr {
		session.Stderr = newAnsiWriter(os.Stderr)
	} else {
		session.Stderr = os.Stderr
	}
	return closeFunc, nil
}

type consoleCaps struct {
	vtInput       bool
	EmulateStdin  bool
	EmulateStdout bool
	EmulateStderr bool
}

const (
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms683167(v=vs.85).aspx
	enableVirtualTerminalInput      = 0x0200
	enableVirtualTerminalProcessing = 0x0004
	disableNewlineAutoReturn        = 0x0008
)

func detectConsoleApp() consoleCaps {
	// https://github.com/docker/docker/blob/de0328560b818e86fd3eadc973f90341e5c33498/pkg/term/term_windows.go#L74-L79
	if os.Getenv("ConEmuANSI") == "ON" || os.Getenv("ConsoleZVersion") != "" || os.Getenv("MSYSCON") != "" {
		return consoleCaps{false, true, false, false}
	}
	var caps consoleCaps
	var fd uintptr
	// Detect console feature detection
	fd = os.Stdin.Fd()
	if mode, err := winterm.GetConsoleMode(fd); err == nil {
		// Validate that enableVirtualTerminalInput is supported, but do not set it.
		if err = winterm.SetConsoleMode(fd, mode|enableVirtualTerminalInput); err != nil {
			caps.EmulateStdin = true
		} else {
			caps.vtInput = true
		}
		// Push back to original modes
		winterm.SetConsoleMode(fd, mode)
	}
	fd = os.Stdout.Fd()
	if mode, err := winterm.GetConsoleMode(fd); err == nil {
		// Validate that enableVirtualTerminalInput is supported, but do not set it.
		if err = winterm.SetConsoleMode(fd, mode|enableVirtualTerminalProcessing|disableNewlineAutoReturn); err != nil {
			caps.EmulateStdout = true
		}
		// Push back to original modes
		winterm.SetConsoleMode(fd, mode|enableVirtualTerminalProcessing)
	}
	fd = os.Stderr.Fd()
	if mode, err := winterm.GetConsoleMode(fd); err == nil {
		// Validate that enableVirtualTerminalInput is supported, but do not set it.
		if err = winterm.SetConsoleMode(fd, mode|enableVirtualTerminalProcessing|disableNewlineAutoReturn); err != nil {
			caps.EmulateStderr = true
		}
		// Push back to original modes
		winterm.SetConsoleMode(fd, mode|enableVirtualTerminalProcessing)
	}

	return caps
}

func getWinSize(fd uintptr) (int, int, error) {
	info, err := winterm.GetConsoleScreenBufferInfo(fd)
	if err != nil {
		return 0, 0, errors.Wrap(err, "winterm.GetConsoleScreenBufferInfo")
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

func makeRaw(fd uintptr) (uint32, error) {
	saved, err := winterm.GetConsoleMode(fd)
	if err != nil {
		return 0, errors.Wrap(err, "Failed winterm.GetConsoleMode")
	}

	// Disable
	raw := saved &^ (winterm.ENABLE_ECHO_INPUT |
		winterm.ENABLE_PROCESSED_INPUT |
		winterm.ENABLE_LINE_INPUT |
		winterm.ENABLE_PROCESSED_OUTPUT)
	// Enable
	raw |= (winterm.ENABLE_WINDOW_INPUT |
		winterm.ENABLE_MOUSE_INPUT)
	if err := winterm.SetConsoleMode(fd, raw); err != nil {
		return 0, errors.Wrap(err, "Failed winterm.SetConsoleMode")
	}
	return saved, nil
}

func restore(fd uintptr, modes uint32) error {
	return errors.Wrap(winterm.SetConsoleMode(fd, modes), "Failed winterm.SetConsoleMode")
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

// Stolen from github.com/docker/docker/pkg/term/windows/ansi_reader.go

func newAnsiReader(in *os.File) io.ReadCloser {
	return &ansiReader{
		file: in,
	}
}

type ansiReader struct {
	file *os.File
}

// Close closes the wrapped file.
func (r *ansiReader) Close() error {
	return r.file.Close()
}

func (r *ansiReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return r.file.Read(p)
}
