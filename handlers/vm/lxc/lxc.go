package lxc

import (
	"encoding/json"
	"io"

	"flag"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
	"github.com/golang/protobuf/proto"
)

func init() {
	handlers.RegisterHandler(&LxcHandler{})
}

type LxcHandler struct {
}

func (h *LxcHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error) {
	tmpl := &model.LxcTemplate{}
	if err := json.Unmarshal(in, tmpl); err != nil {
		return nil, err
	}

	// Validation
	if tmpl.GetLxcImage() == nil {
		return nil, handlers.ErrInvalidTemplate(h, "lxc_image must exist")
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
