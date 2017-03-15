package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var rebootCmd = &cobra.Command{
	Use:   "reboot [Instance ID]",
	Short: "Reboots an instance",
	Long:  "Reboots an instance",
	Example: `
	% openvdc reboot i-xxxxxxx
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			log.Fatalf("Please provide an Instance ID.")
		}

		instanceID := args[0]

		req := &api.RebootRequest{
			InstanceId: instanceID,
		}
		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)
			res, err := c.Reboot(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			fmt.Println(res)
			return err
		})
	},
}
