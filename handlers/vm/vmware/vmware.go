package vmware

import (
	"encoding/json"
	"flag"
	"io"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/handlers/vm"
	"github.com/axsh/openvdc/model"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func init() {
	handlers.RegisterHandler(&VmwareHandler{})
}

type VmwareHandler struct {
	vm.Base
}

func (h *VmwareHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error) {
	tmpl := &model.VmwareTemplate{}
	if err := json.Unmarshal(in, tmpl); err != nil {
		return nil, err
	}
	var json_template struct {
		Template map[string]json.RawMessage `json:"vmware_template,omitempty"`
	}
	if err := json.Unmarshal(in, &json_template); err != nil {
		return nil, err
	}

	return tmpl, nil
}

func (h *VmwareHandler) SetTemplateItem(t *model.Template, m model.ResourceTemplate) {
	t.Item = &model.Template_Vmware{
		Vmware: m.(*model.VmwareTemplate),
	}
}

func (h *VmwareHandler) Merge(dst, src model.ResourceTemplate) error {
	mdst, ok := dst.(*model.VmwareTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.VmwareTemplate), dst)
	}
	msrc, ok := src.(*model.VmwareTemplate)
	if !ok {
		return handlers.ErrMergeSrcType(new(model.VmwareTemplate), src)
	}
	proto.Merge(mdst, msrc)
	return nil
}

func (h *VmwareHandler) MergeArgs(dst model.ResourceTemplate, args []string) error {
	mdst, ok := dst.(*model.VmwareTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.VmwareTemplate), dst)
	}

	flags := flag.NewFlagSet("vmware template", flag.ContinueOnError)
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

func (h *VmwareHandler) Usage(out io.Writer) error {
	return nil
}

func (h *VmwareHandler) MergeJSON(dst model.ResourceTemplate, in json.RawMessage) error {
	mdst, ok := dst.(*model.VmwareTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.VmwareTemplate), dst)
	}
	minput := &model.VmwareTemplate{}
	if err := json.Unmarshal(in, minput); err != nil {
		return errors.WithStack(err)
	}

	proto.Merge(mdst, minput)
	return nil
}
