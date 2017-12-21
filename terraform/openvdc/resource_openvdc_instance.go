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

						"gateway": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

type resourceCallback func() interface{}
type renderCallback func(getResource resourceCallback) ([]byte, error)
type option func () (renderCallback, resourceCallback)

func renderResourceOpt(getResource resourceCallback) ([]byte, error) {
	var buf bytes.Buffer
	i := getResource()
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

func renderInterfaceOpt(getResource resourceCallback) ([]byte, error) {
	// We use a byte buffer because if we'd use a string here, go would create
	// a new string for every concatenation. Not very efficient. :p
	var buf bytes.Buffer

	i := getResource()
	newElement := false
	if x := i; x != nil {
		buf.WriteString("\"interfaces\":[")
		for _, y := range x.([]interface{}) {
			if newElement {
				buf.WriteString(",")
			}

			z := y.(map[string]interface{})
			bytes, err := json.Marshal(z)
			if err != nil {
				return nil, err
			}

			buf.Write(bytes)
			newElement = true
		}
	}
	buf.WriteString("]")
	return buf.Bytes(), nil
}

func renderCmdOpt(options []option) (bytes.Buffer, error) {
	var buf bytes.Buffer
	addOpt := func(fn renderCallback, getResource resourceCallback) error {
		output, err := fn(getResource)
		if err != nil {
			return err
		}
		if len(output) > 0 {
			buf.Write(output)
		}
		return nil
	}

	buf.WriteString("{")
	for idx, opt := range options {
		if idx > 0 {
			buf.WriteString(",")
		}
		renderCb, resourceCb := opt()
		if err := addOpt(renderCb, resourceCb); err != nil {
			return buf, err
		}
	}
	buf.WriteString("}")
	return buf, nil
}

func openVdcInstanceCreate(d *schema.ResourceData, m interface{}) error {
	opts := []option{
		func() (renderCallback, resourceCallback) {
			getResources := func() interface{} {
				return d.Get("resources")
			}
			return renderResourceOpt, getResources
		},
		func() (renderCallback, resourceCallback) {
			getResources := func() interface{} {
				return d.Get("interfaces")
			}
			return renderInterfaceOpt, getResources
		},
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
