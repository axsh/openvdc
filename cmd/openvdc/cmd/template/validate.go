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

var ValidateCmd = &cobra.Command{
	Use:     "validate [Resource Template Path] [template options]",
	Aliases: []string{"test"},
	Short:   "Validate resource template",
	Long:    "Validate resource template",
	Example: `
	% openvdc template validate centos/7/lxc
	% openvdc template validate ./templates/centos/7/null.json
	% openvdc template validate https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json
	`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		templateSlug := args[0]
		rt, err := util.FetchTemplate(templateSlug)
		if err != nil {
			log.Fatal(err)
		}
		rh := rt.Template.ResourceHandler()
		clihn, ok := rh.(handlers.CLIHandler)
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
		return nil
	},
}
