package qemu

import (
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

type Image struct {
	Path    string
	Format  string
	baseImg string
}

type Drive struct {
	Image   *Image
}

func NewImage(format string, baseImage string) (*Image, error) {
	if _, err := os.Stat(baseImage); err != nil {
		return nil, errors.Errorf("File missing: %s", baseImage)
	}
	return &Image{
		Format: format,
		baseImg: baseImage,
	}, nil
}

func (i *Image) CreateInstanceImage(path string) error {
	cmdLine := &cmdLine{args: make([]string, 0)}
	i.Path = path
	cmd := exec.Command("qemu-img", cmdLine.QemuImgCmd(i)...)

	if stdout, err := cmd.CombinedOutput() ; err != nil {
		return errors.Errorf("%s failed with: %s", cmd.Args, stdout)
	}
	return nil
}

func (i *Image) RemoveInstanceImage() error {
	return nil
}
