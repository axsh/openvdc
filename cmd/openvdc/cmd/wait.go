package cmd

import (
	"io"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/axsh/openvdc/model"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var waitCmd = &cobra.Command{
	Use:   "wait [Instance ID] [Instance State]",
	Short: "Wait for the instance until the expected state",
	Long:  "Wait for the instance until the expected state",
	Example: `
	% openvdc wait i-0000000001 running
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) <= 1 {
			log.Fatalf("Missing instance ID and state")
		}

		instanceID := args[0]
		goalStateExpected := args[1]

		goalState, ok := model.InstanceState_State_value[strings.ToUpper(goalStateExpected)]
		if !ok {
			log.Fatalf("Unknown instance state: %s", goalStateExpected)
		}

		req := &api.InstanceEventRequest{
			Target: &api.InstanceIDRequest{
				Key: &api.InstanceIDRequest_ID{
					ID: instanceID,
				},
			},
		}

		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)

			stream, err := c.Event(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Failed Call /api.Instance/Event")
				return err
			}
			for {
				ev, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						break
					} else {
						log.WithError(err).Fatal("Error streaming event")
					}
				}
				if ev.GetState().State == model.InstanceState_State(goalState) {
					return nil
				}
			}
			return err
		})
	},
}
