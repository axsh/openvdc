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

var logCmd = &cobra.Command{
	Use:   "log [ResourceTemplate ID/URI]",
	Short: "Get logs of an instance",
	Long:  "Get logs of an instance",
	Example: `
	% openvdc log i-xxxxxxx
	`,
	DisableFlagParsing: true,
	PreRunE:            util.PreRunHelpFlagCheckAndQuit,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
                        log.Fatalf("Please provide an Instance ID.")
                }

                instanceID := args[0]

                req := &api.LogRequest{
                        InstanceId: instanceID,
                }
                return util.RemoteCall(func(conn *grpc.ClientConn) error {
                        c := api.NewInstanceClient(conn)
                        res, err := c.Log(context.Background(), req)
                        if err != nil {
                                log.WithError(err).Fatal("Disconnected abnormally")
                                return err
                        }
                        fmt.Println(res)
                        return err
                })
	}}
