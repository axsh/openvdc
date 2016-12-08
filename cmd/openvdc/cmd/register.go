package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/axsh/openvdc/registry"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func prepareRegisterAPICall(templateSlug string) *api.ResourceRequest {
	rt, err := util.FetchTemplate(templateSlug)
	if err != nil {
		switch err {
		case registry.ErrUnknownTemplateName:
			log.Fatalf("Template '%s' not found.", templateSlug)
		default:
			log.Fatalf("Invalid path: %s, %v", templateSlug, err)
		}
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
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			log.Fatal("Missing resource file path")
		}

		templateSlug := args[0]
		req := prepareRegisterAPICall(templateSlug)
		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewResourceClient(conn)
			res, err := c.Register(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			fmt.Println(res)
			return err
		})
	},
}
