package cmd

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
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

		resourceID := args[0]
		if len(resourceID) == 0 {
			log.Fatalf("Invalid Resource ID: %s", resourceID)
		}
		req := &api.ResourceIDRequest{
			Key: &api.ResourceIDRequest_ID{
				ID: resourceID,
			},
		}
		return remoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewResourceClient(conn)
			res, err := c.Show(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			/*
				buf, err := json.MarshalIndent(res, "", "  ")
				if err != nil {
					log.WithError(err).Fatal("Faild to format to pretty JSON.")
				}
				fmt.Println(string(buf))
			*/
			m := &proto.TextMarshaler{Compact: false}
			m.Marshal(os.Stdout, res)
			return err
		})
	},
}
