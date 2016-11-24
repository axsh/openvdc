package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var imageName string
var hostName string

func init() {
	// TODO: Remove --server option from sub-command.
	createCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
	createCmd.PersistentFlags().StringVarP(&hostName, "name", "n", "", "Existing host name")
	createCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var createCmd = &cobra.Command{
	Use:   "create [Resource ID]",
	Short: "Create an instance from resource",
	Long:  `Create a new instance from resource.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			log.Fatalf("Please provide a Resource ID.")
		}

		resourceID := args[0]

		req := &api.CreateRequest{
			ResourceId: resourceID,
		}
		return remoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)
			res, err := c.Create(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			fmt.Println(res)
			return err
		})
	},
}
