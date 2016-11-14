package registry

import "errors"

var (
	ErrLocalCacheNotReady = errors.New("Machine image registry has not been cached yet")
	ErrUnknownImageName   = errors.New("Unknown image name")
)

type MachineImageAttribute struct {
	Name        string
	Title       string `json:"Title"`
	Description string `json:"description,omitempty"`
	Images      []struct {
		Hypervisor  string `json:"hypervisor"`
		DownloadURL string `json:"download_url"`
	} `json:"images"`
}

type Registry interface {
	Find(imageName string) (*MachineImageAttribute, error)
	ValidateCache() bool
	IsCacheObsolete() (bool, error)
	Fetch() error
}
