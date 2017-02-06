package cmd

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/axsh/openvdc/model"
	"github.com/shiena/ansicolor"

	"github.com/axsh/openvdc/api"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func sshShell(instanceID string, destAddr string) error {
	config := &ssh.ClientConfig{
		User:    instanceID,
		Timeout: 5 * time.Second,
	}
	conn, err := ssh.Dial("tcp", destAddr, config)
	if err != nil {
		return err
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = ansicolor.NewAnsiColorWriter(os.Stdout)
	session.Stderr = ansicolor.NewAnsiColorWriter(os.Stderr)
	session.Stdin = os.Stdin

	// Handle control + C
	cInt := make(chan os.Signal, 1)
	defer close(cInt)
	signal.Notify(cInt, os.Interrupt)

	fd := int(os.Stdout.Fd())
	if terminal.IsTerminal(fd) {
		w, h, err := terminal.GetSize(fd)
		if err != nil {
			log.WithError(err).Warn("Failed to get console size. Set to 80x40")
			w = 80
			h = 40
		}
		modes := ssh.TerminalModes{
			ssh.ECHO:  0, // Disable echoing
			ssh.IGNCR: 1, // Ignore CR on input.
		}
		if err := session.RequestPty("vt100", h, w, modes); err != nil {
			return err
		}
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
			if err != nil {
				log.WithError(err).Error("Quit")
			}
			return err
		case <-cInt:
			if err := session.Signal(ssh.SIGINT); err != nil {
				log.WithError(err).Error("Failed to send signal")
			}
		}
	}
}

func init() {
	consoleCmd.Flags().Bool("show", false, "Show console information")
}

var consoleCmd = &cobra.Command{
	Use:   "console [Instance ID]",
	Short: "Connect to an instance",
	Long:  "Connect to an instance.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			log.Fatal("Please provide an instance ID")
		}

		instanceID := args[0]

		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			ic := api.NewInstanceClient(conn)
			res, err := ic.Console(context.Background(), &api.ConsoleRequest{InstanceId: instanceID})
			if err != nil {
				log.WithError(err).Fatal("Failed request to Instance.Console API")
			}
			info, err := cmd.Flags().GetBool("show")
			switch res.Type {
			case model.Console_SSH:
				if info {
					host, port, err := net.SplitHostPort(res.GetAddress())
					if err != nil {
						log.Fatal("Invalid ssh host address: ", res.GetAddress())
					}
					fmt.Printf("-p %s %s@%s\n", port, instanceID, host)
					return nil
				}
				if err := sshShell(instanceID, res.GetAddress()); err != nil && err != io.EOF {
					log.WithError(err).Fatal("Failed ssh to ", res.GetAddress())
				}
				return nil
			}
			cc := api.NewInstanceConsoleClient(conn)
			stream, err := cc.Attach(context.Background())
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormally")
				return err
			}
			err = stream.Send(&api.ConsoleIn{InstanceId: instanceID})
			if err != nil {
				return err
			}
			return err
		})
	},
}
