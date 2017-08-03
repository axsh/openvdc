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
	"golang.org/x/crypto/ssh"
)

func init() {
	handlers.RegisterHandler(&LxcHandler{})
}

type LxcHandler struct {
	vm.Base
}

func (h *LxcHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error) {
	tmpl := &model.LxcTemplate{}

	type Download struct {
		Distro  string `json:"distro,omitempty"`
		Release string `json:"release,omitempty"`
	}

	type LxcTemplate struct {
		Download Download `json:"download,omitempty"`
	}

	type LxcImage struct {
		DownloadUrl string `json:"download_url,omitempty"`
		ChksumType  string `json:"chksum_type,omitempty"`
		Chksum      string `json:"chksum,omitempty"`
	}

	// Parse "lxc_template" section if exists.
	var json_template struct {
		LxcTemplate        LxcTemplate `json:"lxc_template,omitempty"`
		LxcImage           LxcImage    `json:"lxc_image,omitempty"`
		AuthenticationType string      `json:"authentication_type,omitempty"`
	}
	if err := json.Unmarshal(in, &json_template); err != nil {
		return nil, err
	}
	if json_template.AuthenticationType != "" {
		format, ok := model.LxcTemplate_AuthenticationType_value[strings.ToUpper(json_template.AuthenticationType)]
		if !ok {
			return nil, errors.Errorf("Unknown value at format: %s", json_template.AuthenticationType)
		}
		tmpl.AuthenticationType = model.LxcTemplate_AuthenticationType(format)

		// Remove authentication_type field
		tmp := make(map[string]interface{})
		var err error
		if err = json.Unmarshal(in, &tmp); err != nil {
			return nil, errors.Wrap(err, "Failed json.Unmarshal")
		}
		delete(tmp, "authentication_type")
		// var err error

		//json_template.AuthenticationType = ""
		in, err = json.Marshal(tmp)

		if err != nil {
			return nil, errors.Wrap(err, "Failed json.Marshal")
		}
	}

	if err := json.Unmarshal(in, tmpl); err != nil {
		return nil, err
	}

	var template struct {
		Template map[string]json.RawMessage `json:"lxc_template,omitempty"`
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

	switch tmpl.AuthenticationType {
	case model.LxcTemplate_NONE:
	case model.LxcTemplate_PUB_KEY:
		if tmpl.SshPublicKey == "" {
			return nil, handlers.ErrInvalidTemplate(h, "ssh_public_key is not set")
		}

		isValidate := validatePublicKey([]byte(tmpl.SshPublicKey))
		if !isValidate {
			return nil, handlers.ErrInvalidTemplate(h, "ssh_public_key is invalid")
		}

	default:
		return nil, handlers.ErrInvalidTemplate(h, "Unknown authentication_type parameter"+tmpl.AuthenticationType.String())
	}
	return tmpl, nil
}

func validatePublicKey(key []byte) bool {
	// Check that the key is in RFC4253 binary format.
	_, err := ssh.ParsePublicKey(key)
	if err == nil {
		return true
	}

	keyStr := string(key[:])
	// Check that the key is in OpenSSH format.
	keyNames := []string{"ssh-rsa", "ssh-dss", "ecdsa-sha2-nistp256", "ssh-ed25519"}
	firstStr := strings.Fields(keyStr)
	for _, name := range keyNames {
		if firstStr[0] == name {
			return true
		}
	}

	// Check that the key is in SECSH format.
	keyNames = []string{"SSH2 ", "RSA", ""}
	for _, name := range keyNames {
		if strings.Contains(keyStr, "---- BEGIN "+name+"PUBLIC KEY ----") &&
			strings.Contains(keyStr, "---- END "+name+"PUBLIC KEY ----") {
			return true
		}
	}
	return false
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

func (h *LxcHandler) MergeJSON(dst model.ResourceTemplate, in json.RawMessage) error {
	mdst, ok := dst.(*model.LxcTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.LxcTemplate), dst)
	}
	minput := &model.LxcTemplate{}
	if err := json.Unmarshal(in, minput); err != nil {
		return errors.WithStack(err)
	}
	// Prevent Image & Template attributes from overwriting.
	minput.LxcImage = nil
	minput.LxcTemplate = nil
	proto.Merge(mdst, minput)
	return nil
}
