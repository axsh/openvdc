// +build terraform

package main

import (
	"github.com/axsh/openvdc/terraform/openvdc"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: openvdc.Provider,
	})
}
