package esxi

import (
	"bytes"
	"testing"

	"github.com/axsh/openvdc/handlers"
	"github.com/axsh/openvdc/model"
	"github.com/stretchr/testify/assert"
)

func TestResourceName(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("vm/esxi", handlers.ResourceName(&EsxiHandler{}))
}

func TestTypes(t *testing.T) {
	assert := assert.New(t)
	assert.Implements((*handlers.ResourceHandler)(nil), &EsxiHandler{})
	assert.Implements((*handlers.CLIHandler)(nil), &EsxiHandler{})
}

const jsonEsxiImage = `{
        "type": "vm/esxi",
        "esxi_image": {
                "download_url": "http://example.com/"
        }
}`

func TestEsxiHandler_ParseTemplate(t *testing.T) {
	assert := assert.New(t)
	h := &EsxiHandler{}
	m, err := h.ParseTemplate(bytes.NewBufferString(jsonEsxiImage).Bytes())
	assert.NoError(err)
	assert.IsType((*model.EsxiTemplate)(nil), m)
	modelesxi := m.(*model.EsxiTemplate)
	assert.NotNil(modelesxi.GetEsxiImage())

	assert.Equal(modelesxi.GetEsxiImage().GetDownloadUrl(), "http://example.com/")
}
