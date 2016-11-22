package registry

import "errors"

var (
	ErrLocalCacheNotReady  = errors.New("Resource template registry has not been cached yet")
	ErrUnknownTemplateName = errors.New("Unknown template name")
)

type MachineImageAttribute struct {
	Title       string `json:"Title"`
	Description string `json:"description,omitempty"`
	Images      []struct {
		Hypervisor      string `json:"hypervisor"`
		DownloadURL     string `json:"download_url"`
		MinimumVCPU     int    `json:"minimum_vcpu"`
		MinimumMemoryGB int    `json:"minimum_memory_gb"`
	} `json:"images"`
}

// Marker interface represents a template item.
func (*MachineImageAttribute) isTemplateDefinition() {}

type isTemplateDefinition interface {
	isTemplateDefinition()
}

type ResourceTemplate struct {
	Name     string
	Template isTemplateDefinition
	remote   *GithubRegistry
}

// Returns absolute URI for the original location of the resource template.
func (r *ResourceTemplate) LocationURI() string {
	return r.remote.LocateURI(r.Name)
}

type Registry interface {
	Find(templateName string) (*ResourceTemplate, error)
	ValidateCache() bool
	IsCacheObsolete() (bool, error)
	Fetch() error
}
