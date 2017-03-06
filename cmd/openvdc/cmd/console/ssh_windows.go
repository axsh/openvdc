// +build windows

package console

import (
	"os"
	"os/signal"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

const CommandExample = `
PS> openvdc console i-000000001
root@i-00000001#
`

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
