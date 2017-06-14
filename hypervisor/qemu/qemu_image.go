
package qemu

import (
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

type Image struct {
	Path    string
	Format  string
	Size    int
	baseImg string
}

type Drive struct {
	Image   *Image
	If      string
}

func NewImage(path string,format string) (*Image) {
	return &Image{
		Format: format,
		Path: path,
	}
}

func (i *Image) SetBaseImage(baseImage string) error {
	// todo check for size ?
	if _, err := os.Stat(baseImage); err != nil {
		return errors.Errorf("File missing: %s", baseImage)
	}
	i.baseImg = baseImage
	return nil
}

func (i *Image) SetSize(size int)  error {
	// todo check for base image ?
	i.Size = size
	return nil
}

func (i *Image) CreateImage() error {
	cmdLine := &cmdLine{args: make([]string, 0)}
	cmd := exec.Command("qemu-img", cmdLine.QemuImgCmd(i)...)
	if stdout, err := cmd.CombinedOutput() ; err != nil {
		return errors.Errorf("%s failed with: %s", cmd.Args, stdout)
	}
	return nil
}

func (i *Image) RemoveInstanceImage() error {
	return nil
}
