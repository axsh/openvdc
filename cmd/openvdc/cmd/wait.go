package cmd

import (
	"io"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/axsh/openvdc/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func init() {
	waitCmd.Flags().String("timeout", "", "Timeout to cancel wait.")
}

var waitCmd = &cobra.Command{
	Use:   "wait [Instance ID] [Instance State]",
	Short: "Wait for the instance until the expected state",
	Long:  "Wait for the instance until the expected state",
	Example: `
	% openvdc wait i-0000000001 running
	% openvdc wait i-0000000001 running --timeout=2h45m
	% openvdc wait i-0000000001 running --timeout=10s
	`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) <= 1 {
			cmd.Usage()
			return errors.New("Missing instance ID and state")
		}

		instanceID := args[0]
		goalStateExpected := args[1]

		waitDuration := 999999 * time.Hour
		if cmd.Flag("timeout").Changed {
			v := cmd.Flag("timeout").Value.String()
			var err error
			waitDuration, err = time.ParseDuration(v)
			if err != nil {
				log.Fatalf("Invalid timeout value: %v", err)
			}
			if waitDuration < 0 {
				log.Fatalf("Timeout value must be positive number")
			}
		}
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
			ctx := context.Background()
			inst, err := c.Show(ctx, req.Target)
			if err != nil {
				log.WithError(err).Fatal("Failed Call /api.Instance/Show")
				return err
			}
			if inst.GetInstance().GetLastState().GetState() == model.InstanceState_State(goalState) {
				// Quit immediatly if the given goal state is identical.
				return nil
			}

			stream, err := c.Event(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Failed Call /api.Instance/Event")
				return err
			}

			quit := make(chan error, 1)

			go func() {
				defer close(quit)
				for {
					ev, err := stream.Recv()
					if err != nil {
						if err == io.EOF {
							return
						}
						log.WithError(err).Error("Error streaming event")
						quit <- err
						return
					}
					if ev.GetState().State == model.InstanceState_State(goalState) {
						quit <- nil
						return
					}
				}
			}()

			tout := time.After(waitDuration)
			for {
				select {
				case err := <-quit:
					return err
				case <-tout:
					quit <- errors.New("Timedout")
				}
			}
			return nil
		})
	},
}
