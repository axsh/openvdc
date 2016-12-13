package template

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/registry"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
)

func mergeTemplateParams(rt *registry.RegistryTemplate, args []string) model.ResourceTemplate {
	if len(args) == 0 {
		return rt.Template.Template
	}

	rh := rt.Template.ResourceHandler()
	clihn, ok := rh.(handlers.CLIHandler)
	if !ok {
		log.Fatal("%s does not support CLI interface", rt.Name)
	}

	pb := proto.Clone(rt.Template.Template.(proto.Message))
	merged, ok := pb.(model.ResourceTemplate)
	if !ok {
		log.Fatal("Failed to cast type")
	}

	subargs := args
	// Process JSON input and merging.
	{
		var err error
		var buf []byte
		if strings.HasPrefix(args[0], "@") {
			fpath := strings.TrimPrefix(args[0], "@")
			buf, err = ioutil.ReadFile(fpath)
			if err != nil {
				log.Fatalf("Failed to read variables from file: %s", fpath)
			}
		} else if args[0] == "-" {
			buf, err = ioutil.ReadAll(os.Stdin)
			if err != nil {
				log.Fatalf("Failed to read variables from stdin")
			}
		} else if !strings.HasPrefix(args[0], "-") {
			// Assume JSON string input
			buf = []byte(args[0])
		}

		if len(buf) > 0 {
			err = json.Unmarshal(buf, merged)
			if err != nil {
				log.Fatal("Invalid variable input:", err)
			}
			subargs = subargs[1:]
		}
	}

	if err := clihn.MergeArgs(merged, subargs); err != nil {
		log.Fatalf("Failed to overwrite parameters for %s\n%s", rt.LocationURI(), err)
	}
	return merged
}

func exampleParameterOverwrite(baseCmd string) string {
	return fmt.Sprintf(`
	Overwrite template parameters:

	Using CLI options:
	%% %[1]s centos/7/lxc --vcpu=4 --memory_gb=4

	Using JSON string:
	%% %[1]s centos/7/lxc '{"vcpu":4, "memory_gb":4}'

	Using variable file:
	%% vi mylxc.json
	{
	  "vcpu": 4,
	  "memory_gb": 4
	}
	%% %[1]s centos/7/lxc @mylxc.json
	%% cat mylxc.json | %[1]s centos/7/lxc -
	`, baseCmd)
}

var ValidateCmd = &cobra.Command{
	Use:     "validate [Resource Template Path] [template options]",
	Aliases: []string{"test"},
	Short:   "Validate resource template",
	Long:    "Validate resource template",
	Example: `
	% openvdc template validate centos/7/lxc
	% openvdc template validate ./templates/centos/7/null.json
	% openvdc template validate https://raw.githubusercontent.com/axsh/openvdc/master/templates/centos/7/lxc.json
	` + exampleParameterOverwrite("openvdc template validate"),
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

		if len(args) > 1 {
			mergeTemplateParams(rt, args[1:])
		}
		return nil
	},
}
