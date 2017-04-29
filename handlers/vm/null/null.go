package null

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/handlers/vm"
	"github.com/axsh/openvdc/model"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func init() {
	handlers.RegisterHandler(&NullHandler{})
}

type NullHandler struct {
	vm.Base
}

func (h *NullHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error) {
	tmpl := &model.NullTemplate{}

	// "CrashStage" is different type between JSON (string) and Protobuf (enum). The "crash_stage"
	// field has to be removed from json.RawMessage if exists.
	var json_template struct {
		CrashStage string `json:"crash_stage,omitempty"`
	}
	if err := json.Unmarshal(in, &json_template); err != nil {
		return nil, errors.Wrap(err, "Failed json.Unmarshal for anonymous struct")
	}
	if json_template.CrashStage != "" {
		crash_stage, ok := model.NullTemplate_CrashStage_value[strings.ToUpper(json_template.CrashStage)]
		if !ok {
			return nil, errors.Errorf("Unknown value at crash_stage: %s", json_template.CrashStage)
		}
		tmpl.CrashStage = model.NullTemplate_CrashStage(crash_stage)

		// Clear "crash_stage" field from "in".
		tmp := make(map[string]interface{})
		if err := json.Unmarshal(in, &tmp); err != nil {
			return nil, errors.Wrap(err, "Failed json.Unmarshal")
		}

		delete(tmp, "crash_stage")
		var err error
		in, err = json.Marshal(tmp)
		if err != nil {
			return nil, errors.Wrap(err, "Failed json.Marshal")
		}
	}

	if err := json.Unmarshal(in, tmpl); err != nil {
		return nil, errors.Wrap(err, "Failed json.Unmarshal for model.NullTemplate")
	}
	return tmpl, nil
}

func (h *NullHandler) SetTemplateItem(t *model.Template, m model.ResourceTemplate) {
	t.Item = &model.Template_Null{
		Null: m.(*model.NullTemplate),
	}
}

func (h *NullHandler) Merge(dst, src model.ResourceTemplate) error {
	mdst, ok := dst.(*model.NullTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.NullTemplate), dst)
	}
	msrc, ok := src.(*model.NullTemplate)
	if !ok {
		return handlers.ErrMergeSrcType(new(model.NullTemplate), src)
	}
	proto.Merge(mdst, msrc)
	return nil
}

func (h *NullHandler) Usage(out io.Writer) error {
	return nil
}

func (h *NullHandler) MergeArgs(src model.ResourceTemplate, args []string) error {
	return nil
}

func (h *NullHandler) MergeJSON(dst model.ResourceTemplate, in json.RawMessage) error {
	mdst, ok := dst.(*model.NullTemplate)
	if !ok {
		return handlers.ErrMergeDstType(new(model.NullTemplate), dst)
	}
	input, err := h.ParseTemplate(in)
	if err != nil {
		return errors.Wrap(err, "Failed to parse input JSON")
	}
	minput, _ := input.(proto.Message)
	proto.Merge(mdst, minput)
	return nil
}
