package template

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show [Resource Template Path]",
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
		return rt.Template.ResourceHandler().ShowHelp(cmd.OutOrStdout())
	},
}
