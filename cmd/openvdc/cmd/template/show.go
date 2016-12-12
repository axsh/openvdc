package template

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show [Resource Template Path] [template options]",
	Short: "Show resource template details",
	Long:  "Show resource template details",
	Example: `
	% openvdc template show centos/7/lxc
	% openvdc template show ./templates/centos/7/null.json
	% openvdc template show https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		templateSlug := args[0]
		rt, err := util.FetchTemplate(templateSlug)
		if err != nil {
			log.Fatal(err)
		}
		clihn, ok := rt.Template.ResourceHandler().(handlers.CLIHandler)
		if !ok {
			return fmt.Errorf("%s does not support CLI interface", rt.Name)
		}
		pb := proto.Clone(rt.Template.Template.(proto.Message))
		merged, ok := pb.(model.ResourceTemplate)
		if !ok {
			return fmt.Errorf("Failed to cast type")
		}

		err = clihn.MergeArgs(merged, args[1:])
		if err != nil {
			log.Fatalf("Failed to overwrite parameters for %s\n%s", rt.LocationURI(), err)
		}

		//TODO: Show difference between rt.ToModel() and merged objects.
		return clihn.Usage(cmd.OutOrStdout())
	},
}
