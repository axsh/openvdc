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
				Type: schema.TypeMap,
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
					},
				},
			},
		},
	}
}

func renderResourceParam(resources interface{}) (bytes.Buffer, error) {
	var buf bytes.Buffer
	if resources != nil {
		bytes, err := json.Marshal(resources.(map[string]interface{}))
		if err != nil {
			return buf, err
		}
		buf.Write(bytes)
	}

	return buf, nil
}

func renderInterfaceParam(nics interface{}) (bytes.Buffer, error) {
	// We use a byte buffer because if we'd use a string here, go would create
	// a new string for every concatenation. Not very efficient. :p
	var buf bytes.Buffer
	buf.WriteString("{\"interfaces\":[")

	newElement := false
	if x := nics; x != nil {
		for _, y := range x.([]interface{}) {
			if newElement {
				buf.WriteString(",")
			}

			z := y.(map[string]interface{})
			bytes, err := json.Marshal(z)
			if err != nil {
				return buf, err
			}

			buf.Write(bytes)
			newElement = true
		}
	}
	return buf, nil
}

func openVdcInstanceCreate(d *schema.ResourceData, m interface{}) error {
	nics, err := renderInterfaceParam(d.Get("interfaces"))
	if err != nil {
		return err
	}
	resources, err := renderResourceParam(d.Get("resources"))
	if err != nil {
		return err
	}

	stdout, stderr, err := RunCmd("openvdc", "run", d.Get("template").(string), nics.String())
	if err != nil {
		return fmt.Errorf("The following command returned error:%v\nopenvdc run %s %s\nSTDOUT: %s\nSTDERR: %s", err, d.Get("template").(string), nics.String(), stdout, stderr)
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
