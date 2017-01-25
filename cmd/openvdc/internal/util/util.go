// Place unexported functions used under cmd namespaces.

package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/registry"
	"google.golang.org/grpc"
)

var UserConfDir string

type Settings struct {
	Interfaces []struct {
		Type string `json:"type"`
	}
}

func init() {
	// UserConfDir variable is referenced from init() in cmd/root.go
	// so that it has to be initialized eagerly.
	//http://stackoverflow.com/questions/7922270/obtain-users-home-directory
	path, err := homedir.Dir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to locate user home path: %v", err)
		os.Exit(1)
	}
	UserConfDir = filepath.Join(path, ".openvdc")
}

func RemoteCall(c func(*grpc.ClientConn) error) error {
	serverAddr := viper.GetString("api.endpoint")
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.WithField("endpoint", serverAddr).Fatalf("Cannot connect to OpenVDC gRPC endpoint: %v", err)
	}
	defer conn.Close()
	return c(conn)
}

func SetupGithubRegistryCache() (registry.TemplateFinder, error) {
	reg := registry.NewGithubRegistry(UserConfDir)
	if !reg.ValidateCache() {
		log.Infoln("Updating registry cache from", reg)
		err := reg.Fetch()
		if err != nil {
			return nil, err
		}
	}

	refresh, err := reg.IsCacheObsolete()
	if err != nil {
		return nil, err
	}
	if refresh {
		log.Infoln("Updating registry cache from", reg)
		err = reg.Fetch()
		if err != nil {
			return nil, err
		}
	}
	return reg, nil
}

func FetchTemplate(templateSlug string) (*registry.RegistryTemplate, error) {
	var finder registry.TemplateFinder
	if strings.HasSuffix(templateSlug, ".json") {
		u, err := url.Parse(templateSlug)
		if err != nil {
			return nil, err
		}
		if u.IsAbs() {
			finder = registry.NewRemoteRegistry()
		} else {
			// Assume the local path string is given.
			finder = registry.NewLocalRegistry()
		}
	} else {
		var err error
		finder, err = SetupGithubRegistryCache()
		if err != nil {
			return nil, err
		}
	}
	return finder.Find(templateSlug)
}

// PreRunHelpFlagCheckAndQuit can use cobra.Command with "DisableFlagParsing=true"
// to show usage and quit for -h or --help command line option.
// &cobra.Command {
//   DisableFlagParsing: true,
// 	 PreRunE:            util.PreRunHelpFlagCheckAndQuit,
// }
func PreRunHelpFlagCheckAndQuit(cmd *cobra.Command, args []string) error {
	cmd.Flags().Parse(args)
	help, _ := cmd.Flags().GetBool("help")
	if help {
		return pflag.ErrHelp
	}
	return nil
}

// MergeTemplateParams returns the value merged resource template. The value source is
// read from JSON string, file, stdin or command line options.
func MergeTemplateParams(rt *registry.RegistryTemplate, args []string) model.ResourceTemplate {
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

			var settings Settings

			err := json.Unmarshal([]byte(buf), &settings)

			if err != nil {
				fmt.Println("Failed reading config file:", err)
			}

			for _, i := range settings.Interfaces {
				if i.Type == "linux" || i.Type == "ovs" {
					continue
				} else {
					log.Fatalf("Unrecognized bridgetype: %s", i.Type)
				}
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

// ExampleMergeTemplateOptions returns the example for command line options
// proceeded by MergeTemplateParams().
func ExampleMergeTemplateOptions(baseCmd string) string {
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
