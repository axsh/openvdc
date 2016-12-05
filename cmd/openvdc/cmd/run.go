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

var runCmd = &cobra.Command{
	Use:   "run [ResourceTemplate ID/URI]",
	Short: "Run an instance",
	Long:  "Run an instance",
	Example: `
	% openvdc run centos/7/lxc
	% openvdc run https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			log.Fatal("Missing resource file path")
		}

		templateSlug := args[0]
		req := prepareRegisterAPICall(templateSlug)
		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)
			res, err := c.Run(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			fmt.Println(res)
			return err
		})
	}}
