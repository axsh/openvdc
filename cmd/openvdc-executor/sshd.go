package main

import (
	"crypto"
	"fmt"
	"io"
	"net"
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"golang.org/x/net/context"
)

type HypervisorProviderFinder interface {
	GetHypervisorProvider() hypervisor.HypervisorProvider
}

type SSHServer struct {
	config   *ssh.ServerConfig
	listener net.Listener
	finder   HypervisorProviderFinder
	ctx      context.Context
}

func NewSSHServer(finder HypervisorProviderFinder, ctx context.Context) *SSHServer {
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	return &SSHServer{
		config: config,
		finder: finder,
		ctx:    ctx,
	}
}

type HostKeyGen func(rand io.Reader) (crypto.Signer, error)

var HostRsaKeyPath string
var HostEcdsaKeyPath string
var HostEd25519KeyPath string

var KeyGenPathList = []string{
	HostRsaKeyPath,
	HostEcdsaKeyPath,
	HostEd25519KeyPath,
}
func (sshd *SSHServer) Setup() error {
	if model.GetBackendCtx(sshd.ctx) == nil {
		return errors.New("Context does not have model connection")
	}
	for _, path := range KeyGenPathList {
		// Reading key file
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, path + " doesn't exist")
		}
		// Check integrity of pem file
		sshSigner, err := ssh.ParsePrivateKey(buf)
		if err != nil {
			return errors.Wrap(err, path + " is not a valid pem file")
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
			instanceID := sshConn.User()

			// TODO: Retrieve the model at the authentication callback.
			inst, err := model.Instances(sshd.ctx).FindByID(instanceID)
			if err != nil {
				log.WithError(err).Errorf("Unknown instance: %s", instanceID)
				sshConn.Close()
				return
			}

			for newChannel := range chans {
				if t := newChannel.ChannelType(); t != "session" {
					newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
					continue
				}
				provider := sshd.finder.GetHypervisorProvider()
				if provider == nil {
					log.Error("HypervisorProvider is not ready")
					newChannel.Reject(ssh.Prohibited, "HypervisorProvider is not ready")
					continue
				}
				hv, err := provider.CreateDriver(inst, inst.ResourceTemplate())
				if err != nil {
					log.WithError(err).Errorf("Failed HypervisorProvider.CreateDriver: %T", provider)
					sshConn.Close()
					return
				}
				session := sshSession{
					instanceID: instanceID,
					sshd:       sshd,
					peer:       sshConn.RemoteAddr(),
					driver:     hv,
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
	driver     hypervisor.HypervisorDriver
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

	console := session.driver.InstanceConsole()
	quit := make(chan error, 1)
	defer close(quit)

Done:
	for {
		select {
		case r := <-req:
			if r == nil {
				quit <- errors.New("Session request is nil")
				break
			}
			log := log.WithField("sshreq", r.Type)
			reply := true
			switch r.Type {
			case "shell":
				ptycon, ok := console.(hypervisor.PtyConsole)
				if session.ptyreq != nil && ok {
					_, err := ptycon.AttachPty(session.console, session.ptyreq)
					if err != nil {
						reply = false
						log.WithError(err).Error("Failed console.AttachPty")
						break
					}
				} else {
					_, err := console.Attach(session.console)
					if err != nil {
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
			case "window-change":
				winchMsg := struct {
					Columns uint32
					Rows    uint32
					Width   uint32
					Height  uint32
				}{}
				if err := ssh.Unmarshal(r.Payload, &winchMsg); err != nil {
					log.WithError(errors.WithStack(err)).Error("Failed to parse window-change request body")
					reply = false
					break
				}
				ptycon, ok := console.(hypervisor.PtyConsole)
				if session.ptyreq != nil && ok {
					if err := ptycon.UpdateWindowSize(winchMsg.Columns, winchMsg.Rows); err != nil {
						log.WithError(err).Error("Failed UpdateWindowSize")
						reply = false
						break
					}
				} else {
					log.Warn("window-change sshreq for non-tty session")
				}
			case "exec":
				var execMsg struct {
					Command string
				}
				if err := ssh.Unmarshal(r.Payload, &execMsg); err != nil {
					log.WithError(errors.WithStack(err)).Error("Failed to parse exec request body")
					reply = false
					break
				}

				// TODO: Skip /bin/sh -c if .Command does not contain shell keywords.
				if _, err := console.Exec(session.console, []string{"/bin/sh", "-c", execMsg.Command}); err != nil {
					log.WithError(err).Error("Failed console.Exec")
					reply = false
					break
				}
				go func() {
					quit <- console.Wait()
				}()
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

			if exiterr, ok := err.(hypervisor.ConsoleWaitError); ok {
				sendExitStatus(uint32(exiterr.ExitCode()))
			} else if err != nil {
				log.WithError(err).Error("Unknown Error")
			} else {
				// err == nil
				sendExitStatus(0)
			}
			break Done
		}
	}
}
