// +build terraform

package openvdc

import (
	"bytes"
	"encoding/json"
	"os/exec"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_endpoint": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OPENVDC_API_ENDPOINT", nil),
				Description: "Endpoint URL for API.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"openvdc_instance": OpenVdcInstance(),
		},

		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
			return config{
				apiEndpoint: d.Get("api_endpoint").(string),
			}, nil
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

func ParseJson(cmdStdOut *bytes.Buffer) (string, error) {
	var parsed map[string]interface{}

	err := json.Unmarshal(cmdStdOut.Bytes(), &parsed)
	if err != nil {
		return "", err
	}
	var x map[string]interface{}
	x = parsed["instance"].(map[string]interface{})

	x = x["last_state"].(map[string]interface{})

	return x["state"].(string), nil

}

func CheckInstanceTerminatedOrFailed(cmdStdOut *bytes.Buffer) (bool, error) {
	status, err := ParseJson(cmdStdOut)

	if err != nil {
		return false, err
	}

	if status == "TERMINATED" || status == "FAILED" {
		return true, nil
	}

	return false, nil

}
