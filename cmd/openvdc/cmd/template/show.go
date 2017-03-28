package template

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	ShowCmd.Flags().String("format", "json", "Output format: (json, protobuf)")
}

var ShowCmd = &cobra.Command{
	Use:   "show [Resource Template Path] [template options]",
	Short: "Show resource template details",
	Long:  "Show resource template details",
	Example: `
	% openvdc template show centos/7/lxc
	% openvdc template show ./templates/centos/7/null.json
	% openvdc template show https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json
	` + util.ExampleMergeTemplateOptions("openvdc template show"),
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

		merged := rt.Template.Template
		if len(args) > 1 {
			merged = util.MergeTemplateParams(rt, args[1:])
		}

		fmt_type, err := cmd.Flags().GetString("format")
		switch fmt_type {
		case "json":
			if err := (&jsonpb.Marshaler{Indent: "  "}).Marshal(cmd.OutOrStdout(), merged.(proto.Message)); err != nil {
				log.WithError(err).Fatal("Faild to format to pretty JSON.")
			}
		case "protobuf":
			//TODO: Show difference between rt.ToModel() and merged objects.
			m := &proto.TextMarshaler{Compact: false}
			return m.Marshal(cmd.OutOrStdout(), merged.(proto.Message))
		default:
			log.Fatalf("Unknown format: %s", fmt_type)
		}
		return nil
	},
}
