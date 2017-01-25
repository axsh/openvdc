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

func prepareRegisterAPICall(templateSlug string, args []string) *api.ResourceRequest {
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
	} else {
		log.Fatalf("Please provide a config file.")
	}
	req := &api.ResourceRequest{
		Template: rt.ToModel(),
	}
	log.Printf("Found template: %s", templateSlug)
	return req
}

var registerCmd = &cobra.Command{
	Use:   "register [ResourceTemplate.json]",
	Short: "Register new resource.",
	Long:  "Register new resource from resource template.",
	Example: `
	% openvdc register centos/7/lxc
	% openvdc register ./templates/centos/7/null.json
	% openvdc register https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json
	` + util.ExampleMergeTemplateOptions("openvdc register"),
	DisableFlagParsing: true,
	PreRunE:            util.PreRunHelpFlagCheckAndQuit,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return pflag.ErrHelp
		}

		templateSlug := args[0]
		req := prepareRegisterAPICall(templateSlug, args)
		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewResourceClient(conn)
			res, err := c.Register(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			fmt.Println(res.GetID())
			return err
		})
	},
}
