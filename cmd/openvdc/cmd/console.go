package cmd

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/cmd/console"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/axsh/openvdc/model"

	"github.com/axsh/openvdc/api"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func init() {
	consoleCmd.Flags().Bool("show", false, "Show console information")
}

var consoleCmd = &cobra.Command{
	Use:   "console [Instance ID] [options] [--] [commands]",
	Short: "Connect to an instance",
	Long:  "Connect to an instance.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			log.Fatal("Please provide an instance ID")
		}

		instanceID := args[0]
		execArgs := []string{}
		if len(args) > 1 {
			for _, a := range args[1:] {
				if a == "--" {
					// Ignore args before "--"
					execArgs = []string{}
				}
				execArgs = append(execArgs, a)
			}
		}

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
				fmt.Printf("-p %s %s@%s", port, instanceID, host)
				if len(execArgs) > 0 {
					fmt.Printf(" %s", strings.Join(execArgs, " "))
				}
				fmt.Println("")
				return nil
			}
			sshcon := console.NewSshConsole(instanceID, nil)
			var err error
			if len(execArgs) > 0 {
				err = sshcon.Exec(res.GetAddress(), execArgs)
			} else {
				err = sshcon.Run(res.GetAddress())
			}
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
