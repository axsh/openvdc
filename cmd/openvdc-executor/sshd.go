package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net"
	"runtime"
	"syscall"

	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/pkg/errors"
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

type HostKeyGen func(rand io.Reader) (crypto.Signer, error)

var KeyGenList = []HostKeyGen{
	func(rand io.Reader) (crypto.Signer, error) {
		_, priv, err := ed25519.GenerateKey(rand)
		return priv, err
	},
	func(rand io.Reader) (crypto.Signer, error) {
		return ecdsa.GenerateKey(elliptic.P521(), rand)
	},
	func(rand io.Reader) (crypto.Signer, error) {
		return rsa.GenerateKey(rand, 2048)
	},
}

func (sshd *SSHServer) Setup() error {
	for _, gen := range KeyGenList {
		priv, err := gen(rand.Reader)
		if err != nil {
			return errors.Wrap(err, "Failed to generate host key")
		}
		sshSigner, err := ssh.NewSignerFromSigner(priv)
		if err != nil {
			return errors.Wrap(err, "Failed to convert to ssh.Signer")
		}
		sshd.config.AddHostKey(sshSigner)
	}
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
		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		go ssh.DiscardRequests(reqs)
		go func() {
			for newChannel := range chans {
				if t := newChannel.ChannelType(); t != "session" {
					newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
					continue
				}
				session := sshSession{
					instanceID: sshConn.User(),
					sshd:       sshd,
					peer:       sshConn.RemoteAddr(),
				}
				go session.handleChannel(newChannel)
			}
		}()
	}
}

type sshSession struct {
	instanceID string
	sshd       *SSHServer
	peer       net.Addr
	ptyreq     *hypervisor.SSHPtyReq
	console    *hypervisor.ConsoleParam
}

func (session *sshSession) handleChannel(newChannel ssh.NewChannel) {
	log := log.WithField("instance_id", session.instanceID).WithField("peer", session.peer.String())
	connection, req, err := newChannel.Accept()
	if err != nil {
		log.Error("Could not accept channel:", err)
		return
	}
	session.console = hypervisor.NewConsoleParam(connection, connection, connection.Stderr())
	defer func() {
		if err := connection.CloseWrite(); err != nil && err != io.EOF {
			log.WithError(err).Warn("Failed CloseWrite()")
		}
		if err := connection.Close(); err != nil && err != io.EOF {
			log.WithError(err).Warn("Invalid close sequence")
		}
		log.Info("Session closed")
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
			log := log.WithField("sshreq", r.Type)
			reply := true
			switch r.Type {
			case "shell":
				ptycon, ok := console.(hypervisor.PtyConsole)
				if session.ptyreq != nil && ok {
					if err := ptycon.AttachPty(session.console, session.ptyreq); err != nil {
						reply = false
						log.WithError(err).Error("Failed console.AttachPty")
						break
					}
				} else {
					if err := console.Attach(session.console); err != nil {
						reply = false
						log.WithError(err).Error("Failed console.Attach")
						break
					}
				}
				go func() {
					err := console.Wait()
					log.WithError(err).Info("Console released")
					quit <- err
				}()
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
				ptyreq := new(hypervisor.SSHPtyReq)
				if err := ssh.Unmarshal(r.Payload, ptyreq); err != nil {
					reply = false
					log.WithError(errors.WithStack(err)).Error("Failed to parse message")
				} else {
					session.ptyreq = ptyreq
				}
			case "env":
				var envReq struct {
					Name  string
					Value string
				}
				if err := ssh.Unmarshal(r.Payload, &envReq); err != nil {
					log.WithError(errors.WithStack(err)).Error("Failed to parse env request body")
					reply = false
					break
				}
				session.console.Envs[envReq.Name] = envReq.Value
			default:
				reply = false
				log.Warn("Unsupported session request")
			}

			if r.WantReply {
				if err := r.Reply(reply, nil); err != nil {
					log.WithError(errors.WithStack(err)).Warn("Failed to reply")
				}
			}
		case err := <-quit:
			sendExitStatus := func(code uint32) {
				msg := struct {
					ExitStatus uint32
				}{uint32(code)}
				_, err := connection.SendRequest("exit-status", false, ssh.Marshal(&msg))
				if err != nil {
					log.WithError(err).Error("Failed to send exit-status")
				} else {
					log.WithField("exit-status", msg.ExitStatus).Info("Reply exit-status")
				}
			}

			if exiterr, ok := err.(*exec.ExitError); ok {
				// http://stackoverflow.com/questions/10385551/get-exit-code-go
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					sendExitStatus(uint32(status.ExitStatus()))
				} else {
					log.Warnf("This platform %s does not support syscall.WaitStatus: %v", runtime.GOOS, exiterr)
					if exiterr.Success() {
						sendExitStatus(0)
					} else {
						sendExitStatus(1)
					}
				}
			} else if err != nil {
				log.WithError(err).Error("")
			} else {
				// err == nil
				sendExitStatus(0)
			}
			break Done
		}
	}
}
