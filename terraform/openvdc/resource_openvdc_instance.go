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

func openVdcInstanceCreate(d *schema.ResourceData, m interface{}) error {
	// We use a byte buffer because if we'd use a string here, go would create
	// a new string for every concatenation. Not very efficient. :p
	var cmdOpts bytes.Buffer

	cmdOpts.WriteString("{\"interfaces\":[")
	newElement := false
	if x := d.Get("interfaces"); x != nil {
		for _, y := range x.([]interface{}) {
			if newElement {
				cmdOpts.WriteString(",")
			}

			z := y.(map[string]interface{})
			bytes, err := json.Marshal(z)
			if err != nil {
				return err
			}

			cmdOpts.Write(bytes)
			newElement = true
		}
	}
	cmdOpts.WriteString("]}")

	stdout, stderr, err := RunCmd("openvdc", "run", d.Get("template").(string), cmdOpts.String())
	if err != nil {
		return fmt.Errorf("The following command returned error:%v\nopenvdc run %s %s\nSTDOUT: %s\nSTDERR: %s", err, d.Get("template").(string), cmdOpts.String(), stdout, stderr)
	}

	d.SetId(strings.TrimSpace(stdout.String()))

	return nil
}

func openVdcInstanceDelete(d *schema.ResourceData, m interface{}) error {
	stdout, stderr, err := RunCmd("openvdc", "destroy", d.Id())

	if err != nil {
		return fmt.Errorf("The following command returned error:%v\nopenvdc destroy %s\nSTDOUT: %s\nSTDERR: %s", err, d.Id(), stdout, stderr)
	}

	return nil
}

//TODO: Never ever use this again
func notImplemented(d *schema.ResourceData, m interface{}) error {
	return nil
}
