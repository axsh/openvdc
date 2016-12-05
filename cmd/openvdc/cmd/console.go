package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/axsh/openvdc/api"
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

                return remoteCall(func(conn *grpc.ClientConn) error {
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
