package cmd

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/axsh/openvdc/model"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var forceStateCmd = &cobra.Command{
	Use:   "force-state [Instance ID] [State]",
	Short: "Forcefully sets instance to specified state",
	Long:  "Forcefully sets instance to specified state",
	Example: `
	% openvdc force-state i-0000000001 running
	`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			log.Fatalf("Please provide an instance ID.")
		}

		if len(args) == 1 {
			log.Fatalf("Please provide a desired instance state.")
		}

		instanceID := args[0]

		if instanceID == "" {
			log.Fatalf("Invalid Instance ID")
		}

		state := strings.ToUpper(args[1])

		goalState, ok := model.InstanceState_State_value[state]
		if !ok {
			log.Fatalf("Unknown instance state: %s", state)
		}

		req := &api.ForceStateRequest{
			InstanceId: instanceID,
			State:      model.InstanceState_State(goalState),
		}

		var res *api.ForceStateReply

		err := util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)
			var err error
			res, err = c.ForceState(context.Background(), req)
			return err
		})

		if err != nil {
			log.WithError(err).Fatal("Disconnected abnormally")
		}

		return nil
	},
}
