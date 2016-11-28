// Place unexported functions used under cmd namespaces.

package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/registry"
	"google.golang.org/grpc"
)

func remoteCall(c func(*grpc.ClientConn) error) error {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.WithField("endpoint", serverAddr).Fatalf("Cannot connect to OpenVDC gRPC endpoint: %v", err)
	}
	defer conn.Close()
	return c(conn)
}

func setupGithubRegistryCache() (registry.TemplateFinder, error) {
	reg := registry.NewGithubRegistry(UserConfDir)
	if !reg.ValidateCache() {
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
		log.Infoln("Updating registry cache.")
		err = reg.Fetch()
		if err != nil {
			return nil, err
		}
	}
	return reg, nil
}
