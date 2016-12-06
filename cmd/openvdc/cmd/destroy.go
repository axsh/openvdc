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

var destroyCmd = &cobra.Command{
	Use:   "destroy [Instance ID]",
	Short: "Destroy an instance",
	Long:  "Destroy an already existing instance.",
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) < 1 {
			log.Fatal("Please provide an instance ID")
		}

		instanceID := args[0]

		req := &api.DestroyRequest{
			InstanceId: instanceID,
		}

		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)

			res, err := c.Destroy(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormally")
				return err
			}
			fmt.Println(res)
			return err
		})
	},
}
