package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

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

var indentityFile string

func init() {
	consoleCmd.Flags().Bool("show", false, "Show console information")
	consoleCmd.Flags().StringVarP(&indentityFile, "identity-file", "i", "", "Selects a file from which the identity (private key) for public key authentication is read")
}

var consoleCmd = &cobra.Command{
	Use:     "console [Instance ID] [options] [--] [commands]",
	Short:   "Connect to an instance",
	Long:    "Connect to an instance.",
	Example: console.CommandExample,
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

		info, _ := cmd.Flags().GetBool("show")
		fmt.Printf("show is %v", info)

		err := util.RemoteCall(func(conn *grpc.ClientConn) error {
			ic := api.NewInstanceClient(conn)
			var err error
			res, err = ic.Console(context.Background(), &api.ConsoleRequest{InstanceId: instanceID})
			return err
		})
		if err != nil {
			log.WithError(err).Fatal("Failed request to Instance.Console API")
		}

		// info, err := cmd.Flags().GetBool("show")
		// fmt.Printf("show is %v", info)
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

			var config *ssh.ClientConfig
			// Parse and set indetifyFifle
			if indentityFile != "" {
				key, err := ioutil.ReadFile(indentityFile)
				if err != nil {
					log.Fatalf("unable to read private key: %v", err)
				}

				// Create the Signer for this private key.
				signer, err := ssh.ParsePrivateKey(key)
				if err != nil {
					log.Fatalf("unable to parse private key: %v", err)
				}

				config = &ssh.ClientConfig{
					Timeout: 5 * time.Second,
					Auth: []ssh.AuthMethod{
						ssh.PublicKeys(signer),
					},
				}
			} else {
				config = &ssh.ClientConfig{
					Timeout: 5 * time.Second,
					Auth: []ssh.AuthMethod{
						ssh.Password(""),
					},
				}
			}

			sshcon := console.NewSshConsole(instanceID, config)
			// var err error
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
