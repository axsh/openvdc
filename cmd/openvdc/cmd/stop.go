package cmd

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
)

var stopCmd = &cobra.Command{
	Use:   "stop [Instance ID]",
	Short: "Stop a running instance",
	Long:  "Stop a running instance.",
	Example: "openvdc stop i-0000000001",
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) != 1 {
			log.Fatalf("Please provide an instance ID.")
		}

		instanceID := args[0]

		req := &api.StopRequest{
			InstanceId: instanceID,
		}

		return util.RemoteCall(func(conn *grpc.ClientConn) error {
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
