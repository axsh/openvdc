package openvdc

import (
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
		},
	}
}

func openVdcInstanceCreate(d *schema.ResourceData, m interface{}) error {
	stdout, _, err := RunCmd("openvdc", "run", d.Get("template").(string))
	if err != nil {
		return err
	}

	d.SetId(strings.TrimSpace(stdout.String()))

	return nil
}

func openVdcInstanceDelete(d *schema.ResourceData, m interface{}) error {
	_, _, err := RunCmd("openvdc", "destroy", d.Id())

	return err
}

//TODO: Never ever use this again
func notImplemented(d *schema.ResourceData, m interface{}) error {
	return nil
}
