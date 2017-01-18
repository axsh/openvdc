package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)


var bridgeType string
          
func init() {
	runCmd.Flags().StringVarP(&bridgeType, "bridge_type", "t", "", "Bridge type")
}

var runCmd = &cobra.Command{
	Use:   "run [ResourceTemplate ID/URI]",
	Short: "Run an instance",
	Long:  "Run an instance",
	Example: `
	% openvdc run centos/7/lxc
	% openvdc run https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json
	` + util.ExampleMergeTemplateOptions("openvdc run"),
	DisableFlagParsing: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := util.PreRunHelpFlagCheckAndQuit(cmd, args)
		if err != nil {
			return err
		}
		err = cmd.ParseFlags(args)

		if err != nil {
			fmt.Println(err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
			
		left := cmd.Flags().Args()
		if len(left) < 1 {
			return pflag.ErrHelp
		}
	
		if util.IsFlagProvided(args,"bridge_type","t") != true {
			log.Fatalf("Please specify a bridge type.")
		}

		args = util.HandleArgs(args)
		
		templateSlug := left[0]
		for i, a := range args {
			
			if a == templateSlug {
				left = args[i:]
				break
			}
		}

		req := prepareRegisterAPICall(templateSlug, left)
		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)
			res, err := c.Run(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			fmt.Println(res.GetInstanceId())
			return err
		})
	}}
