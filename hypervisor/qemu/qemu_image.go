package qemu

import (
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

type Drive struct {
	Path   string
	Format string
}

type Image struct {
	Path    string
	Format  string
	baseImg string
}

func NewImage(path string, format string, baseImage string) (*Image, error) {
	if _, err := os.Stat(baseImage); err != nil {
		return nil, errors.Errorf("File missing: %s", baseImage)
	}
	return &Image{
		Path: path,
		Format: format,
		baseImg: baseImage,
	}, nil
}

func (i *Image) CreateInstanceImage() error {
	cmdLine := &cmdLine{args: make([]string, 0)}
	cmd := exec.Command("qemu-img", cmdLine.qemuImgCmd(i)...)

	if stdout, err := cmd.CombinedOutput() ; err != nil {
		return errors.Errorf("%s failed with: %s", cmd.Args, stdout)
	}
	return nil
}
