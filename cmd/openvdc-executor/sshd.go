package main

import (
	"fmt"
	"io"
	"net"

	"github.com/pkg/errors"

	"github.com/axsh/openvdc/hypervisor"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
)

type SSHServer struct {
	config   *ssh.ServerConfig
	listener net.Listener
	provider hypervisor.HypervisorProvider
}

func NewSSHServer(provider hypervisor.HypervisorProvider) *SSHServer {
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	return &SSHServer{
		config:   config,
		provider: provider,
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
		go sshd.handleChannels(chans, instanceID)
	}
}

func (sshd *SSHServer) handleChannels(chans <-chan ssh.NewChannel, instanceID string) {
	for newChannel := range chans {
		if t := newChannel.ChannelType(); t != "session" {
			newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			continue
		}
		session := sshSession{instanceID: instanceID, sshd: sshd}
		go session.handleChannel(newChannel)
	}
}

type sshSession struct {
	instanceID string
	sshd       *SSHServer
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
		if err := connection.CloseWrite(); err != nil && err != io.EOF {
			log.WithError(err).Warn("Failed CloseWrite()")
		}
		if err := connection.Close(); err != nil && err != io.EOF {
			log.WithError(err).Warn("Invalid close sequence")
		}
		log.WithField("instance_id", session.instanceID).Info("Session closed")
	}()

	driver, err := session.sshd.provider.CreateDriver(session.instanceID)
	if err != nil {
		log.Error(err)
		return
	}
	console := driver.InstanceConsole()
	quit := make(chan error, 1)
	defer close(quit)

Done:
	for {
		select {
		case r := <-req:
			if r == nil {
				break Done
			}
			switch r.Type {
			case "shell":
				if err := console.Attach(connection, connection, connection.Stderr()); err != nil {
					log.Error(err)
					return
				}

				go func() {
					quit <- console.Wait()
				}()
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
				case ssh.SIGINT, ssh.SIGKILL:
					console.ForceClose()
				default:
					log.Warn("FIXME: Uncovered signal request: ", msg.Signal)
				}
			case "pty-req":
				var ptyReq struct {
					Term     string
					Columns  uint32
					Rows     uint32
					Width    uint32
					Height   uint32
					Modelist string
				}
				if err := ssh.Unmarshal(r.Payload, &ptyReq); err != nil {
					log.WithError(err).Error("Failed to parse pty-req message")
				}
				if err := r.Reply(true, nil); err != nil {
					log.WithError(err).Warn("Failed to reply")
				}
			default:
				if r.WantReply {
					r.Reply(false, nil)
				}
				log.Warn("Unsupported session request: ", r.Type)
			}
		case err := <-quit:
			if err != nil {
				log.WithError(err).Error("")
			}
			break Done
		}
	}
}