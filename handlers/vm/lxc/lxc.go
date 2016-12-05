package lxc

import (
	"encoding/json"
	"io"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
)

func init() {
	handlers.RegisterHandler("vm/lxc", &LxcHandler{})
}

type LxcHandler struct {
}

func (h *LxcHandler) ParseTemplate(in json.RawMessage) (model.ResourceTemplate, error) {
	return &model.LxcTemplate{}, nil
}

func (h *LxcHandler) ShowHelp(out io.Writer) error {
	return nil
}
