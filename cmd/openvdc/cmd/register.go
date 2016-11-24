package cmd

import (
	"fmt"

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
		// TODO: handle direct content URI parameter.

		reg, err := setupLocalRegistry()
		if err != nil {
			log.Fatalln(err)
		}
		rt, err := reg.Find(templateSlug)
		if err != nil {
			if err == registry.ErrUnknownTemplateName {
				log.Fatalf("Template '%s' not found.", templateSlug)
			} else {
				log.Fatalln(err)
			}
		}
		log.Printf("Found template: %s", templateSlug)
		req := &api.ResourceRequest{
			Template: &api.ResourceRequest_Vm{
				Vm: &model.VMTemplate{
					Vcpu:             1,
					MemoryGb:         1,
					ImageTemplateUri: rt.LocationURI(),
				},
			},
		}
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
