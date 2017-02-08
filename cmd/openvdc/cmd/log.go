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

var tail bool

func init() {
	logCmd.Flags().BoolVarP(&tail, "tail", "t", false, "Tail log output instead of just printing it once.")
}

var logCmd = &cobra.Command{
	Use:   "log [Instance ID]",
	Short: "Print logs of an instance",
	Long:  "Print logs of an instance",
	Example: `
	% openvdc log i-xxxxxxx
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			log.Fatalf("Please provide an Instance ID.")
		}
		instanceID := args[0]
		req := &api.InstanceLogRequest{
			Target: &api.InstanceIDRequest{
				Key: &api.InstanceIDRequest_ID{
					ID: instanceID,
				},
			},
		}
		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)
			res, err := c.Log(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			for _, l := range res.Line {
				fmt.Println(l)
			}
			return nil
		})
	},
}
