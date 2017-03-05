package cmd

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
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

const defaultTermInfo = "vt100"

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

	session.Stdin = os.Stdin

	// Handle control + C
	cInt := make(chan os.Signal, 1)
	defer close(cInt)
	signal.Notify(cInt, os.Interrupt, syscall.SIGWINCH)

	fd := int(os.Stdin.Fd())
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
			switch sig {
			case syscall.SIGWINCH:
				w, h, err := terminal.GetSize(fd)
				if err != nil {
					log.WithError(err).Error("Failed terminal.GetSize")
					break
				}
				winchMsg := struct {
					Columns uint32
					Rows    uint32
					Width   uint32
					Height  uint32
				}{
					Columns: uint32(w),
					Rows:    uint32(h),
					Width:   uint32(w * 8),
					Height:  uint32(h * 8),
				}
				if _, err := session.SendRequest("window-change", false, ssh.Marshal(&winchMsg)); err != nil {
					log.WithError(err).Error("Failed session.SendRequest(window-change)")
					break
				}
			case os.Interrupt:
				sshSig := ssh.SIGINT
				if err := session.Signal(sshSig); err != nil {
					log.WithError(err).Error("Failed to send signal")
				}
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

		var res *api.ConsoleReply
		err := util.RemoteCall(func(conn *grpc.ClientConn) error {
			ic := api.NewInstanceClient(conn)
			var err error
			res, err = ic.Console(context.Background(), &api.ConsoleRequest{InstanceId: instanceID})
			return err
		})
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
			err := sshShell(instanceID, res.GetAddress())
			switch err.(type) {
			case *ssh.ExitError:
				defer os.Exit(err.(*ssh.ExitError).ExitStatus())
			case *ssh.ExitMissingError:
				log.Fatal(err.Error())
			case nil:
				// Nothing to do
			default:
				if err == io.EOF {
					return nil
				}
				log.WithError(err).Fatal("Failed ssh to ", res.GetAddress())
			}
			return nil
		default:
			log.Fatalf("Unsupported console type: %s", res.Type.String())
		}
		/*
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
		*/
		return nil
	},
}
