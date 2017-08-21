package lxc

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/handlers/vm"
	"github.com/axsh/openvdc/model"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func init() {
	handlers.RegisterHandler(&LxcHandler{})
}

type LxcHandler struct {
	vm.Base
}

func (h *LxcHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error) {
	var template struct {
		Template map[string]json.RawMessage `json:"lxc_template,omitempty"`
	}
	tmpl := &model.LxcTemplate{}
	in, authType, err := h.Base.ValidateAuthenticationType(in)
	if err != nil {
		return nil, err
	}
	tmpl.AuthenticationType = authType

	if err := json.Unmarshal(in, tmpl); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(in, &template); err != nil {
		return nil, err
	}
	if template.Template != nil {
		if len(template.Template) != 1 {
			return nil, fmt.Errorf("lxc_template section must contain one JSON object")
		}
		// Take only head item
		for k, raw := range template.Template {
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

	err = h.Base.ValidatePublicKey(h, tmpl.AuthenticationType, tmpl.SshPublicKey)
	if err != nil {
		return nil, err
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
	var authType, sshPubkey string
	flags.IntVar(&vcpu, "vcpu", int(mdst.MinVcpu), "")
	flags.IntVar(&mem, "memory_gb", int(mdst.MinMemoryGb), "")
	defAuth := model.AuthenticationType_name[int32(mdst.AuthenticationType)]
	flags.StringVar(&authType, "authentication_type", defAuth, "")
	flags.StringVar(&sshPubkey, "ssh_public_key", mdst.SshPublicKey, "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	mdst.Vcpu = int32(vcpu)
	mdst.MemoryGb = int32(mem)
	format := model.AuthenticationType_value[strings.ToUpper(authType)]
	mdst.AuthenticationType = model.AuthenticationType(format)
	sshPubkey = strings.Replace(sshPubkey, "\"", "", -1)
	mdst.SshPublicKey = sshPubkey
	return nil
}

func (h *LxcHandler) Usage(out io.Writer) error {
	return nil
}

func (h *LxcHandler) MergeJSON(dst model.ResourceTemplate, in json.RawMessage) error {
	mdst, ok := dst.(*model.LxcTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.LxcTemplate), dst)
	}
	minput := &model.LxcTemplate{}
	in, authType, err := h.Base.ValidateAuthenticationType(in)
	if err != nil {
		return errors.WithStack(err)
	}
	minput.AuthenticationType = authType

	if err := json.Unmarshal(in, minput); err != nil {
		return errors.WithStack(err)
	}

	err = h.Base.ValidatePublicKey(h, minput.AuthenticationType, minput.SshPublicKey)
	if err != nil {
		return err
	}

	// Prevent Image & Template attributes from overwriting.
	minput.LxcImage = nil
	minput.LxcTemplate = nil
	proto.Merge(mdst, minput)
	return nil
}
