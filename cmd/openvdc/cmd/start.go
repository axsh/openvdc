package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var startCmd = &cobra.Command{
	Use:   "start [Instance ID]",
	Short: "Start an instance",
	Long:  `Start an instance in REGISTERED or STOPPED state to RUNNING.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			log.Fatalf("Please provide an Instance ID.")
		}

		instanceID := args[0]

		req := &api.StartRequest{
			InstanceId: instanceID,
		}
		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)
			_, err := c.Start(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			return err
		})
	}}
