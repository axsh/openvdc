package cmd

import (
	"fmt"
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/registry"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func prepareRegisterAPICall(templateSlug string) *api.ResourceRequest {
	var finder registry.TemplateFinder
	if strings.HasSuffix(templateSlug, ".json") {
		u, err := url.Parse(templateSlug)
		if err != nil {
			log.Fatal("Invalid path: ", templateSlug)
		}
		if u.IsAbs() {
			finder = registry.NewRemoteRegistry()
		} else {
			// Assume the local path string is given.
			finder = registry.NewLocalRegistry()
		}
	} else {
		var err error
		finder, err = setupGithubRegistryCache()
		if err != nil {
			log.Fatalln(err)
		}
	}
	rt, err := finder.Find(templateSlug)
	if err != nil {
		if err == registry.ErrUnknownTemplateName {
			log.Fatalf("Template '%s' not found.", templateSlug)
		} else {
			log.Fatalln(err)
		}
	}
	req := &api.ResourceRequest{
		TemplateUri: rt.LocationURI(),
	}
	// TODO: Define the factory method.
	{
		t := rt.Template.Template
		switch t.(type) {
		case *model.NullTemplate:
			req.Template = &api.ResourceRequest_Null{
				Null: t.(*model.NullTemplate),
			}
		case *model.LxcTemplate:
			req.Template = &api.ResourceRequest_Lxc{
				Lxc: t.(*model.LxcTemplate),
			}
		}
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
		return remoteCall(func(conn *grpc.ClientConn) error {
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
