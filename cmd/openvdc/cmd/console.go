package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/axsh/openvdc/model"

	"github.com/axsh/openvdc/api"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

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
				return err
			}
			switch res.Type {
			case model.Console_SSH:
				fmt.Printf("ssh %s\n", res.GetAddress())
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
