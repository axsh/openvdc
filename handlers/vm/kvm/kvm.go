package lxc

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
	handlers.RegisterHandler(&KvmHandler{})
}

type KvmHandler struct {
	vm.Base
}

func (h *LxcHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error){
	return nil, nil
}

func (h *KvmHandler) SetTemplateItem(t *model.Template, m model.ResourceTemplate) {
	t.Item = &model.Template_Kvm{
		Kvm: m.(*model.KvmTemplate),
	}
}

func (h *KvmHandler) Merge(dst, src model.ResourceTemplate) error {
	mdst, ok := dst.(*model.KvmTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.KvmTemplate), dst)
	}
	msrc, ok := src.(*model.KvmTemplate)
	if !ok {
		return handlers.ErrMergeSrcType(new(model.KvmTemplate), src)
	}
	proto.Merge(mdst, msrc)
	return nil
}

func (h *KvmHandler) MergeArgs(dst model.ResourceTemplate, args []string) error {
	mdst, ok := dst.(*model.KvmTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.KvmTemplate), dst)
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

func (h *KvmHandler) Usage(out io.Writer) error {
	return nil
}

func (h *KvmHandler) MergeJSON(dst model.ResourceTemplate, in json.RawMessage) error {
	mdst, ok := dst.(*model.KvmTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.KvmTemplate), dst)
	}
	minput := &model.KvmTemplate{}
	if err := json.Unmarshal(in, minput); err != nil {
		return errors.WithStack(err)
	}
	// Prevent Image & Template attributes from overwriting.
	minput.KvmImage = nil
	proto.Merge(mdst, minput)
	return nil
}
