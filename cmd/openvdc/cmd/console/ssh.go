package console

import (
	"os"
	"os/signal"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
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
		return errors.Wrap(err, "Failed ssh.Dial")
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return errors.Wrap(err, "Failed ssh.NewSession")
	}
	defer session.Close()

	// Handle control + C
	cInt := make(chan os.Signal, 1)
	defer close(cInt)
	s.signal(cInt)

	closeFunc, err := s.bindFDs(session)
	if err != nil {
		return err
	}
	defer func() {
		err := closeFunc()
		if err != nil {
			log.WithError(err).Error("Failed close process")
		}
	}()

	if err := session.Shell(); err != nil {
		return errors.Wrap(err, "Failed session.Shell")
	}

	quit := make(chan error, 1)
	defer close(quit)

	go func() {
		quit <- session.Wait()
	}()

	for {
		select {
		case err := <-quit:
			return errors.WithStack(err)
		case sig := <-cInt:
			err := s.signalHandle(sig, session)
			if err != nil {
				log.WithError(err).Error("Failed signalHandle")
				quit <- err
			}
		}
	}
}

func (s *SshConsole) Exec(destAddr string, args []string) error {
	s.ClientConfig.User = s.instanceID
	conn, err := ssh.Dial("tcp", destAddr, s.ClientConfig)
	if err != nil {
		return errors.Wrap(err, "Failed ssh.Dial")
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return errors.Wrap(err, "Failed ssh.NewSession")
	}
	defer session.Close()

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// Handle control + C only
	cInt := make(chan os.Signal, 1)
	defer close(cInt)
	signal.Notify(cInt, os.Interrupt)

	if err := session.Start(strings.Join(args, " ")); err != nil {
		return errors.Wrap(err, "Failed session.SendRequest(exec)")
	}

	quit := make(chan error, 1)
	defer close(quit)

	go func() {
		quit <- session.Wait()
	}()

	for {
		select {
		case err := <-quit:
			return errors.WithStack(err)
		case sig := <-cInt:
			err := s.signalHandle(sig, session)
			if err != nil {
				log.WithError(err).Error("Failed signalHandle")
				quit <- err
			}
		}
	}
}
