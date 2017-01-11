package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"

	"github.com/axsh/openvdc/api"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var consoleCmd = &cobra.Command{
	Use:   "console [Instance ID]",
	Short: "Connect to an instance",
	Long:  "Connect to an instance.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {

                        log.Fatal("Please provide an instance ID")
                }

                instanceID := args[0]

                req := &api.ConsoleRequest{
                        InstanceId: instanceID,
                }

                return util.RemoteCall(func(conn *grpc.ClientConn) error {
                        c := api.NewInstanceClient(conn)

                        res, err := c.Console(context.Background(), req)
                        if err != nil {
                                log.WithError(err).Fatal("Disconnected abnormally")
                                return err
                        }
                        fmt.Println(res)
                        return err
                })
	},
}
