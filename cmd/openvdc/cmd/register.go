package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/registry"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func init() {
	// TODO: Remove --server option from sub-command.
	registerCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
	registerCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

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
	log.Printf("Found template: %s", templateSlug)
	return &api.ResourceRequest{
		TemplateUri: rt.LocationURI(),
		Template: &api.ResourceRequest_Lxc{
			Lxc: &model.LxcTemplate{
				Vcpu:     1,
				MemoryGb: 1,
			},
		},
	}
}

var registerCmd = &cobra.Command{
	Use:   "register [ResourceTemplate.json]",
	Short: "Register new resource.",
	Long:  "Register new resource from resource template.",
	Example: `
	% openvdc register centos-7
	% openvdc register https://raw.githubusercontent.com/axsh/openvdc-images/master/centos-7.json
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
