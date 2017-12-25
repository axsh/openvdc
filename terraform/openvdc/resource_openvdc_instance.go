// +build terraform

package openvdc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"strings"
)

func OpenVdcInstance() *schema.Resource {
	return &schema.Resource{
		Create: openVdcInstanceCreate,
		Read:   notImplemented,
		Update: notImplemented,
		Delete: openVdcInstanceDelete,

		Schema: map[string]*schema.Schema{

			"template": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"resources": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vcpu": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
						},
						"memory_gb": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},

			"interfaces": &schema.Schema{
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "veth",
						},

						"bridge": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"ipv4addr": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"macaddr": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"ipv4gateway": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

type option func() ([]byte, error)

func renderMapOpt(i interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if i != nil {
		x := i.(map[string]interface{})
		bytes, err := json.Marshal(x)
		if err != nil {
			return nil, err
		}
		buf.Write(bytes)
	}

	var b []byte
	b = bytes.Trim(buf.Bytes(), "{")
	b = bytes.Trim(b, "}")
	return b, nil
}

func renderListOpt(param string, i interface{}) ([]byte, error) {
	// We use a byte buffer because if we'd use a string here, go would create
	// a new string for every concatenation. Not very efficient. :p
	var buf bytes.Buffer
	newElement := false

	if x := i; x != nil {
		buf.WriteString(strings.Join([]string{"\"", param, "\":[{"}, ""))
		for _, y := range x.([]interface{}) {
			if newElement {
				buf.WriteString("},{")
			}
			bytes, err := renderMapOpt(y)
			if err != nil {
				return nil, err
			}

			buf.Write(bytes)
			newElement = true
		}
	}
	buf.WriteString("}]")
	return buf.Bytes(), nil
}

func renderCmdOpt(options []option) (bytes.Buffer, error) {
	var buf bytes.Buffer

	buf.WriteString("{")
	for idx, optCb := range options {
		if idx > 0 {
			buf.WriteString(",")
		}
		output, err := optCb()
		if err != nil {
			return buf, err
		}
		if len(output) > 0 {
			buf.Write(output)
		}
	}
	buf.WriteString("}")
	return buf, nil
}

func openVdcInstanceCreate(d *schema.ResourceData, m interface{}) error {
	opts := []option{
		func() ([]byte, error) { return renderMapOpt(d.Get("resources")) },
		func() ([]byte, error) { return renderListOpt("interfaces", d.Get("interfaces"))},
	}

	cmdOpts, err := renderCmdOpt(opts)
	if err != nil {
		return err
	}

	stdout, stderr, err := RunCmd("openvdc", "run", d.Get("template").(string), cmdOpts.String())
	if err != nil {
		return fmt.Errorf("The following command returned error:%v\nopenvdc run %s %s\nSTDOUT: %s\nSTDERR: %s", err, d.Get("template").(string), cmdOpts.String(), stdout, stderr)
	}

	d.SetId(strings.TrimSpace(stdout.String()))

	return nil
}

func openVdcInstanceDelete(d *schema.ResourceData, m interface{}) error {
	stdout, stderr, err := RunCmd("openvdc", "show", d.Id())
	if err != nil {
		return fmt.Errorf("The following command returned error:%v\nopenvdc show %s\nSTDOUT: %s\nSTDERR: %s", err, d.Id(), stdout, stderr)
	}
	instanceAlreadyTerminated, err := CheckInstanceTerminatedOrFailed(stdout)

	if err != nil {
		return fmt.Errorf("Error parsing json output for openvdc show command for id %s. Error: %s", d.Id(), err)
	}

	if instanceAlreadyTerminated {
		return nil
	}

	stdout, stderr, err = RunCmd("openvdc", "destroy", d.Id())

	if err != nil {
		return fmt.Errorf("The following command returned error:%v\nopenvdc destroy %s\nSTDOUT: %s\nSTDERR: %s", err, d.Id(), stdout, stderr)
	}

	return nil
}

//TODO: Never ever use this again
func notImplemented(d *schema.ResourceData, m interface{}) error {
	return nil
}
