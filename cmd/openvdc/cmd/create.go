package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/axsh/openvdc/registry"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var hostName string

func init() {
	createCmd.PersistentFlags().StringVarP(&hostName, "name", "n", "", "Existing host name")
}

func prepareCreateAPICall(templateSlug string, args []string) *api.CreateRequest {
	rt, err := util.FetchTemplate(templateSlug)
	if err != nil {
		switch err {
		case registry.ErrUnknownTemplateName:
			log.Fatalf("Template '%s' not found.", templateSlug)
		default:
			log.Fatalf("Invalid path: %s, %v", templateSlug, err)
		}
	}
	if len(args) > 1 {
		rt.Template.Template = util.MergeTemplateParams(rt, args[1:])
	}
	req := &api.CreateRequest{
		Template: rt.ToModel(),
	}
	log.Printf("Found template: %s", templateSlug)
	return req
}

var createCmd = &cobra.Command{
	Use:   "create [ResourceTemplate.json]",
	Short: "Register an instance from template",
	Long:  `Register an instance from template.`,
	Example: `
	% openvdc create centos/7/lxc
	% openvdc create ./templates/centos/7/null.json
	% openvdc create https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json
	` + util.ExampleMergeTemplateOptions("openvdc create"),
	DisableFlagParsing: true,
	PreRunE:            util.PreRunHelpFlagCheckAndQuit,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return pflag.ErrHelp
		}

		templateSlug := args[0]
		req := prepareCreateAPICall(templateSlug, args)
		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)
			res, err := c.Create(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			fmt.Println(res.GetInstanceId())
			return err
		})
	},
}
