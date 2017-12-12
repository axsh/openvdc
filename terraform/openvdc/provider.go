package openvdc

import (
	"bytes"
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

func CheckInstanceTerminated(cmdStdOut *bytes.Buffer) (value bool) {
     for _, s := range (strings.Split(cmdStdOut.String(), "\n"))  {
        if has:=strings.Contains(s, "\"state\""); has {
             vals := strings.Split(s, ":")
             _, state := vals[0], vals[1]
             return strings.Contains(state, "TERMINATED")
         }
    }
    return false

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
