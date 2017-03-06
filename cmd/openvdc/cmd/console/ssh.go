package console

import (
	"os"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/shiena/ansicolor"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

const defaultTermInfo = "vt100"

type SshConsole struct {
	instanceID   string
	ClientConfig *ssh.ClientConfig
}

func NewSshConsole(instanceID string, config *ssh.ClientConfig) *SshConsole {
	if config == nil {
		config = &ssh.ClientConfig{
			Timeout: 5 * time.Second,
		}
	}
	return &SshConsole{
		instanceID:   instanceID,
		ClientConfig: config,
	}
}

func (s *SshConsole) Run(destAddr string) error {
	s.ClientConfig.User = s.instanceID
	conn, err := ssh.Dial("tcp", destAddr, s.ClientConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = os.Stdin

	// Handle control + C
	cInt := make(chan os.Signal, 1)
	defer close(cInt)
	s.signal(cInt)

	fd := int(os.Stdin.Fd())
	if terminal.IsTerminal(fd) {
		w, h, err := terminal.GetSize(fd)
		if err != nil {
			log.WithError(err).Warn("Failed to get console size. Set to 80x40")
			w = 80
			h = 40
		}
		modes := ssh.TerminalModes{}
		term, ok := os.LookupEnv("TERM")
		if !ok {
			term = defaultTermInfo
		}
		if err := session.RequestPty(term, h, w, modes); err != nil {
			return err
		}

		origstate, err := terminal.MakeRaw(fd)
		if err != nil {
			return err
		}
		defer func() {
			if err := terminal.Restore(fd, origstate); err != nil {
				if errno, ok := err.(syscall.Errno); (ok && errno != 0) || !ok {
					log.WithError(err).Error("Failed terminal.Restore")
				}
			}
		}()
		session.Stdout = ansicolor.NewAnsiColorWriter(os.Stdout)
		session.Stderr = ansicolor.NewAnsiColorWriter(os.Stderr)
	} else {
		session.Stdout = os.Stdout
		session.Stderr = os.Stderr
	}

	if err := session.Shell(); err != nil {
		return err
	}

	quit := make(chan error, 1)
	defer close(quit)

	go func() {
		quit <- session.Wait()
	}()

	for {
		select {
		case err := <-quit:
			return err
		case sig := <-cInt:
			err := s.signalHandle(sig, session)
			if err != nil {
				log.WithError(err).Error("Failed signalHandle")
				quit <- err
			}
		}
	}
}
