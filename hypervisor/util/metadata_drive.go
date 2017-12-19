package util

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type MetadataDrive interface {
	MetadataDrivePath() string
	MetadataDriveDatamap() map[string]interface{}
}

func runCmd(cmd string, args []string) error {
	c := exec.Command(cmd, args...)

	if err := c.Run(); err != nil {
		return errors.Wrapf(err, "failed to execute command :%s %s", cmd, args)
	}
	return nil
}

func renderData(keyPath string, key string, value interface{}) error {
	switch value.(type) {
	case string:
		if err := ioutil.WriteFile(filepath.Join(keyPath, key), []byte(value.(string)), 0644); err != nil {
			return errors.Wrapf(err, "failed to write to file: %s", filepath.Join(keyPath, key))
		}
	default:
		kp := filepath.Join(keyPath, key)
		if err := os.MkdirAll(kp, os.ModePerm); err != nil {
			return errors.Wrapf(err, "Unable to create folder: %s", kp)
		}
		for key, value := range value.(map[string]interface{}) {
			if err := renderData(kp, key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func MountMetadataDisk(md MetadataDrive) error {
	mountPath := filepath.Join(filepath.Dir(md.MetadataDrivePath()), "meta-data")

	if err := os.MkdirAll(mountPath, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create folder: %s", mountPath)
	}
	if err := runCmd("mount", []string{md.MetadataDrivePath(), mountPath}); err != nil {
		return errors.Wrap(err, "failed to mount metadrive image")
	}
	return nil
}

func UmountMetadataDisk(md MetadataDrive) error {
	mountPath := filepath.Join(filepath.Dir(md.MetadataDrivePath()), "meta-data")

	if err := runCmd("umount", []string{mountPath}); err != nil {
		return errors.Wrap(err, "failed to umount metadrive image")
	}
	if err := os.RemoveAll(mountPath); err != nil {
		return errors.Wrapf(err, "falied to remove folder: %s", mountPath)
	}
	return nil
}

func WriteMetadata(md MetadataDrive) error {
	for key, value := range md.MetadataDriveDatamap() {
		if err := renderData(filepath.Join(filepath.Dir(md.MetadataDrivePath()), "meta-data"), key, value); err != nil {
			return err
		}
	}
	return nil
}

func CreateMetadataDisk(md MetadataDrive) error {
	log.Infoln("Preparing metadrive image...")

	if _, err := os.Stat(md.MetadataDrivePath()); err != nil {
		if err := runCmd("mkfs.msdos", []string{"-C", "-s", "1", md.MetadataDrivePath(), "1440"}); err != nil {
			return errors.Wrap(err, "failed to create metadrive image")
		}
	} else {
		if err := runCmd("mkfs.msdos", []string{"-s", "1", md.MetadataDrivePath()}); err != nil {
			return errors.Wrap(err, "failed to format metadrive image")
		}
	}
	return nil
}
