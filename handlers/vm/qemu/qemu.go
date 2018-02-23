package qemu

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
	handlers.RegisterHandler(&QemuHandler{})
}

type QemuHandler struct {
	vm.Base
}

func (h *QemuHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error) {
	tmpl := &model.QemuTemplate{QemuImage: &model.QemuTemplate_Image{}}

	type QemuImage struct {
		DownloadUrl string `json:"download_url,omitempty"`
		Format      string `json:"format,omitempty"`
	}

	var json_template struct {
		QemuImage QemuImage `json:"qemu_image,omitempty"`
	}

	if err := json.Unmarshal(in, &json_template); err != nil {
		return nil, errors.Wrap(err, "Failed json.Unmarshal for anonymous struct")
	}

	if json_template.QemuImage.Format != "" {
		format, ok := model.QemuTemplate_Image_Format_value[strings.ToUpper(json_template.QemuImage.Format)]
		if !ok {
			return nil, errors.Errorf("Unknown value at format: %s", json_template.QemuImage.Format)
		}
		tmpl.QemuImage.Format = model.QemuTemplate_Image_Format(format)

		// Remove format field
		tmp := make(map[string]interface{})
		if err := json.Unmarshal(in, &tmp); err != nil {
			return nil, errors.Wrap(err, "Failed json.Unmarshal")
		}
		delete(tmp["qemu_image"].(map[string]interface{}), "format")
		var err error
		in, err = json.Marshal(tmp)
		if err != nil {
			return nil, errors.Wrap(err, "Failed json.Marshal")
		}
	}

	in, authType, err := h.Base.ValidateAuthenticationType(in)
	if err != nil {
		return nil, err
	}
	tmpl.AuthenticationType = authType

	if err := json.Unmarshal(in, tmpl); err != nil {
		return nil, errors.Wrap(err, "Failed json.Unmarshal for model.QemuTemplate")
	}

	// Validation
	if tmpl.GetQemuImage() == nil {
		return nil, handlers.ErrInvalidTemplate(h, "qemu_image must exist")
	}

	return tmpl, nil
}

func (h *QemuHandler) SetTemplateItem(t *model.Template, m model.ResourceTemplate) {
	qemuTmpl, ok := m.(*model.QemuTemplate)
	if !ok {
		panic("template type is not *model.QemuTemplate")
	}
	t.Item = &model.Template_Qemu{
		Qemu: qemuTmpl,
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
	format, ok := model.AuthenticationType_value[strings.ToUpper(authType)]
	if !ok {
		return fmt.Errorf("Unknown AuthenticationType: %s", authType)
	}
	mdst.AuthenticationType = model.AuthenticationType(format)
	sshPubkey = strings.Replace(sshPubkey, "\"", "", -1)
	mdst.SshPublicKey = sshPubkey
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
	minput.QemuImage = nil
	proto.Merge(mdst, minput)
	return nil
}
