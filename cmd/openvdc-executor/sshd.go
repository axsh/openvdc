package main

import (
	"fmt"
	"net"

	"github.com/pkg/errors"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
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

		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		go ssh.DiscardRequests(reqs)
		go handleChannels(chans)
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}
	connection, _, err := newChannel.Accept()
	if err != nil {
		log.Error("Could not accept channel:", err)
		return
	}
	defer connection.Close()
}
