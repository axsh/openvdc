package openvdc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"os/exec"
	"strings"
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

func CheckInstanceTerminated(cmdStdOut *bytes.Buffer) (bool, error) {
	status, err := ParseJson(cmdStdOut)

	if err != nil {
		return false, err
	}

	if strings.Compare(status, "TERMINATED") == 0 {
		return true, nil
	}

	return false, nil

}
