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
	// TODO: Remove --server option from sub-command.
	registerCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
	registerCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var registerCmd = &cobra.Command{
	Use:   "register [Resource.json]",
	Short: "Register new resource definition",
	Long:  "Register new resource from resource definition file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			log.Fatal("Missing resource file path")
		}

		return remoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewResourceClient(conn)
			res, err := c.Register(context.Background(), &api.ResourceRequest{})
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			fmt.Println(res)
			return err
		})
	},
}
