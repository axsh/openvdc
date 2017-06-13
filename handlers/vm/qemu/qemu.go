package qemu

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
	handlers.RegisterHandler(&QemuHandler{})
}

type QemuHandler struct {
	vm.Base
}

func (h *QemuHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error) {
	tmpl := &model.QemuTemplate{}
	if err := json.Unmarshal(in, tmpl); err != nil {
		return nil, err
	}

	// Validation
	if tmpl.GetQemuImage() == nil {
		return nil, handlers.ErrInvalidTemplate(h, "qemu_image must exist")
	}

	return tmpl, nil
}

func (h *QemuHandler) SetTemplateItem(t *model.Template, m model.ResourceTemplate) {
	t.Item = &model.Template_Qemu{
		Qemu: m.(*model.QemuTemplate),
	}
}

func (h *QemuHandler) Merge(dst, src model.ResourceTemplate) error {
	mdst, ok := dst.(*model.QemuTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.QemuTemplate), dst)
	}
	msrc, ok := src.(*model.QemuTemplate)
	if !ok {
		return handlers.ErrMergeSrcType(new(model.QemuTemplate), src)
	}
	proto.Merge(mdst, msrc)
	return nil
}

func (h *QemuHandler) MergeArgs(dst model.ResourceTemplate, args []string) error {
	mdst, ok := dst.(*model.QemuTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.QemuTemplate), dst)
	}

	flags := flag.NewFlagSet("qemu template", flag.ContinueOnError)
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

func (h *QemuHandler) Usage(out io.Writer) error {
	return nil
}

func (h *QemuHandler) MergeJSON(dst model.ResourceTemplate, in json.RawMessage) error {
	mdst, ok := dst.(*model.QemuTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.QemuTemplate), dst)
	}
	minput := &model.QemuTemplate{}
	if err := json.Unmarshal(in, minput); err != nil {
		return errors.WithStack(err)
	}
	// Prevent Image & Template attributes from overwriting.
	minput.QemuImage = nil
	proto.Merge(mdst, minput)
	return nil
}
