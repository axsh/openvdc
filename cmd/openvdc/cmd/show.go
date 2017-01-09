package cmd

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var showCmd = &cobra.Command{
	Use:   "show [Resource ID]",
	Short: "Show a resource",
	Long:  `Show a resource.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			log.Fatalf("Please provide a Resource ID.")
		}

		type showTargetCb func(conn *grpc.ClientConn) (proto.Message, error)
		var showTarget showTargetCb
		id := args[0]
		if strings.HasPrefix(id, "i-") {
			showTarget = func(conn *grpc.ClientConn) (proto.Message, error) {
				req := &api.InstanceIDRequest{
					Key: &api.InstanceIDRequest_ID{
						ID: id,
					},
				}

				c := api.NewInstanceClient(conn)
				return c.Show(context.Background(), req)
			}
		} else if strings.HasPrefix(id, "r-") {
			showTarget = func(conn *grpc.ClientConn) (proto.Message, error) {
				req := &api.ResourceIDRequest{
					Key: &api.ResourceIDRequest_ID{
						ID: id,
					},
				}

				c := api.NewResourceClient(conn)
				return c.Show(context.Background(), req)
			}
		} else {
			log.Fatal("Invalid ID syntax:", id)
		}

		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			res, err := showTarget(conn)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			if err := (&jsonpb.Marshaler{Indent: "  "}).Marshal(os.Stdout, res); err != nil {
				log.WithError(err).Fatal("Faild to format to pretty JSON.")
			}
			return err
		})
	},
}
