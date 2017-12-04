package openvdc

import (
	"bytes"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"os/exec"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},

		ResourcesMap: map[string]*schema.Resource{
			"openvdc_instance": OpenVdcInstance(),
		},
	}

}

func RunCmd(name string, arg ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(name, arg...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return &stdout, &stderr, err
}
