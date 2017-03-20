package lxc

import (
	"encoding/json"
	"fmt"
	"io"

	"flag"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/handlers/vm"
	"github.com/axsh/openvdc/model"
	"github.com/golang/protobuf/proto"
)

func init() {
	handlers.RegisterHandler(&LxcHandler{})
}

type LxcHandler struct {
	vm.Base
}

func (h *LxcHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error) {
	tmpl := &model.LxcTemplate{}
	if err := json.Unmarshal(in, tmpl); err != nil {
		return nil, err
	}

	// Parse "lxc_template" section if exists.
	var json_template struct {
		Template map[string]json.RawMessage `json:"lxc_template,omitempty"`
	}
	if err := json.Unmarshal(in, &json_template); err != nil {
		return nil, err
	}
	if json_template.Template != nil {
		if len(json_template.Template) != 1 {
			return nil, fmt.Errorf("lxc_template section must contain one JSON object")
		}
		// Take only head item
		for k, raw := range json_template.Template {
			tmpl.LxcTemplate = &model.LxcTemplate_Template{
				Template: k,
			}
			if err := json.Unmarshal(raw, tmpl.LxcTemplate); err != nil {
				return nil, err
			}
			break
		}
	}

	// Validation
	if tmpl.GetLxcImage() == nil && tmpl.GetLxcTemplate() == nil {
		return nil, handlers.ErrInvalidTemplate(h, "lxc_image or lxc_template must exist")
	}

	return tmpl, nil
}

func (h *LxcHandler) SetTemplateItem(t *model.Template, m model.ResourceTemplate) {
	t.Item = &model.Template_Lxc{
		Lxc: m.(*model.LxcTemplate),
	}
}

func (h *LxcHandler) Merge(dst, src model.ResourceTemplate) error {
	mdst, ok := dst.(*model.LxcTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.LxcTemplate), dst)
	}
	msrc, ok := src.(*model.LxcTemplate)
	if !ok {
		return handlers.ErrMergeSrcType(new(model.LxcTemplate), src)
	}
	proto.Merge(mdst, msrc)
	return nil
}

func (h *LxcHandler) MergeArgs(dst model.ResourceTemplate, args []string) error {
	mdst, ok := dst.(*model.LxcTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.LxcTemplate), dst)
	}

	flags := flag.NewFlagSet("lxc template", flag.ContinueOnError)
	var vcpu, mem int
	flags.IntVar(&vcpu, "vcpu", int(mdst.MinVcpu), "")
	flags.IntVar(&mem, "memory_gb", int(mdst.MinMemoryGb), "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	mdst.Vcpu = int32(vcpu)
	mdst.MemoryGb = int32(mem)
	return nil
}

func (h *LxcHandler) Usage(out io.Writer) error {
	return nil
}
