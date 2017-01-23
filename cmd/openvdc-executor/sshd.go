package main

import (
	"fmt"
	"io"
	"net"

	"github.com/pkg/errors"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type SSHServer struct {
	config   *ssh.ServerConfig
	listener net.Listener
}

func NewSSHServer() *SSHServer {
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	return &SSHServer{
		config: config,
	}
}

func (sshd *SSHServer) Setup() error {
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return errors.Wrap(err, "Failed to generate host key")
	}
	sshSigner, err := ssh.NewSignerFromSigner(priv)
	if err != nil {
		return errors.Wrap(err, "Failed to convert to ssh.Signer")
	}
	sshd.config.AddHostKey(sshSigner)
	return nil
}

func (sshd *SSHServer) Run(listener net.Listener) {
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Error("Failed to accept incoming connection:", err)
			continue
		}
		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, sshd.config)
		if err != nil {
			log.Error("Failed to handshake:", err)
			continue
		}
		instanceID := sshConn.User()
		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		go ssh.DiscardRequests(reqs)
		go handleChannels(chans, instanceID)
	}
}

func handleChannels(chans <-chan ssh.NewChannel, instanceID string) {
	for newChannel := range chans {
		if t := newChannel.ChannelType(); t != "session" {
			newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			continue
		}
		session := sshSession{instanceID: instanceID}
		go session.handleChannel(newChannel)
	}
}

type sshSession struct {
	instanceID string
}

func (session *sshSession) handleChannel(newChannel ssh.NewChannel) {
	connection, req, err := newChannel.Accept()
	if err != nil {
		log.Error("Could not accept channel:", err)
		return
	}
	defer func() {
		msg := struct {
			ExitStatus uint32
		}{uint32(0)}
		_, err := connection.SendRequest("exit-status", false, ssh.Marshal(&msg))
		if err != nil {
			log.WithError(err).Error("Failed to send exit-status")
		}
		if err := connection.Close(); err != nil {
			log.WithError(err).Warn("Invalid close sequence")
		} else {
			log.Info("Session closed")
		}
	}()

	quit := make(chan error, 1)
	go func(connection ssh.Channel) {
		t := terminal.NewTerminal(connection, "> ")
		var err error
		defer func() {
			quit <- err
			close(quit)
		}()

		for {
			l, err := t.ReadLine()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Error("failed to read: %v", err)
				return
			}

			if _, err := t.Write([]byte("You've typed: " + string(l) + "\r\n")); err != nil {
				log.Printf("failed to write: %v", err)
				return
			}
		}
	}(connection)

Done:
	for {
		select {
		case r := <-req:
			switch r.Type {
			case "shell":
				if err := r.Reply(true, nil); err != nil {
					log.WithError(err).Warn("Failed to reply")
				}

			case "signal":
				var msg struct {
					Signal string
				}
				if err := ssh.Unmarshal(r.Payload, &msg); err != nil {
					log.WithError(err).Error("Failed to parse signal requeyst body")
					// Won't break the loop
					break
				}

				switch ssh.Signal(msg.Signal) {
				case ssh.SIGINT:
					break Done
				default:
					log.Warn("FIXME: Uncovered signal request: ", msg.Signal)
				}
			default:
				if r.WantReply {
					r.Reply(false, nil)
				}
				log.Warn("Unsupported session request: ", r.Type)
			}
		case <-quit:
			break Done
		}
	}
}
