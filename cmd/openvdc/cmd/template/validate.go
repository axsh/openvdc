package template

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/axsh/openvdc/registry"
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

		finder, err := util.TemplateFinder(templateSlug)
		if err != nil {
			log.WithError(err).Fatal("Failed util.TemplateFinder")
		}
		buf, err := finder.LoadRaw(templateSlug)
		if err != nil {
			log.WithError(err).Fatal("Failed finder.LoadRaw")
		}
		if err, ok := registry.ValidateTemplate(buf).(*registry.ErrInvalidTemplate); ok && err != nil {
			log.Fatal(err.Errors)
		}

		if len(args) > 1 {
			rt, err := util.FetchTemplate(templateSlug)
			if err != nil {
				log.WithError(err).Fatal("util.FetchTemplate")
			}
			util.MergeTemplateParams(rt, args[1:])
		}
		return nil
	},
}
