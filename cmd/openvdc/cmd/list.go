package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/axsh/openvdc/model"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List instances",
	Long:  "List instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &api.InstanceListRequest{
			Filter: &api.InstanceListRequest_Filter{
				State: model.InstanceState_REGISTERED,
			},
		}
		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)
			res, err := c.List(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			for i, item := range res.Items {
				fmt.Printf("%-5d %-10s %-10s\n", i, item.Id, item.State.String())
			}
			return err
		})
	}}
