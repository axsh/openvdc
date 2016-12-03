package cmd

import (
	"fmt"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var unregisterCmd = &cobra.Command{
	Use:   "unregister [Resource ID]",
	Short: "Unregister a resource",
	Long:  "Unregister existing resource by specifiying resource ID or resource name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			log.Fatal("Missing resource ID/name")
		}

		resourceID := args[0]
		if len(resourceID) == 0 {
			log.Fatalf("Invalid Resource ID: %s", resourceID)
		}
		return remoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewResourceClient(conn)
			res, err := c.Unregister(context.Background(), &api.ResourceIDRequest{Key: &api.ResourceIDRequest_ID{ID: resourceID}})
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			fmt.Println(res)
			return err
		})
	},
}
