package template

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	` + util.ExampleMergeTemplateOptions("openvdc template validate"),
	DisableFlagParsing: true,
	PreRunE:            util.PreRunHelpFlagCheckAndQuit,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return pflag.ErrHelp
		}
		templateSlug := args[0]
		rt, err := util.FetchTemplate(templateSlug)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) > 1 {
			util.MergeTemplateParams(rt, args[1:])
		}
		return nil
	},
}
