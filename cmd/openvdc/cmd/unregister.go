package cmd

import (
	"fmt"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func init() {
	unregisterCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
	unregisterCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var unregisterCmd = &cobra.Command{
	Use:   "unregister [Resource ID]",
	Short: "Unregister a resource",
	Long:  "Unregister existing resource by specifiying resource ID or resource name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			log.Fatal("Missing resource ID/name")
		}

		return APICall(func(conn *grpc.ClientConn) error {
			c := api.NewResourceClient(conn)
			res, err := c.Unregister(context.Background(), &api.ResourceIDRequest{})
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			fmt.Println(res)
			return err
		})
	},
}
