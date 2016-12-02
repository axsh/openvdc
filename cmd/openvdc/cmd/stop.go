package cmd

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/axsh/openvdc/api"
)

func init() {
	// TODO: Remove --server option from sub-command.
	stopCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
	stopCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var stopCmd = &cobra.Command{
	Use:   "stop [Instance ID]",
	Short: "Stop an instance",
	Long:  "Stop a running instance.",
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) != 1 {
			log.Fatalf("Please provide an instance ID.")
		}

		instanceID := args[0]

		req := &api.StopRequest{
			InstanceId: instanceID,
		}

		return remoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)

			res, err := c.Stop(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormally")
				return err
			}
			fmt.Println(res)
			return err
		})
	},
}
