// Place unexported functions used under cmd namespaces.

package util

import (
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/axsh/openvdc/registry"
	"google.golang.org/grpc"
)

var ServerAddr string
var UserConfDir string

func RemoteCall(c func(*grpc.ClientConn) error) error {
	conn, err := grpc.Dial(ServerAddr, grpc.WithInsecure())
	if err != nil {
		log.WithField("endpoint", ServerAddr).Fatalf("Cannot connect to OpenVDC gRPC endpoint: %v", err)
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
