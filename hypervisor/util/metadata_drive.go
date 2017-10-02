package util

import (
	"os"
	"os/exec"
	"io/ioutil"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type MetadataDrive interface {
	MetadataDrivePath()    string
	MetadataDriveDatamap() map[string]interface{}
}

func runCmd(cmd string, args []string) error {
	c := exec.Command(cmd, args...)
	if err := c.Run(); err != nil {
		return errors.Errorf("failed to execute command :%s %s", cmd, args)
	}
	return nil
}

func renderData(keyPath string, key string, value interface{}) error {
	switch value.(type) {
	case string:
		ioutil.WriteFile(filepath.Join(keyPath, key), []byte(value.(string)), 0644)
		return nil
	default:
		kp := filepath.Join(keyPath, key)
		if err := os.MkdirAll(kp, os.ModePerm); err != nil {
			return errors.Errorf("Unable to create folder: %s", kp)
		}
		for key, value := range value.(map[string]interface{}) {
			if err := renderData(kp, key, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func WriteMetadata(md MetadataDrive) error {
	mountPath := filepath.Join(filepath.Dir(md.MetadataDrivePath()), "meta-data")

	if err := os.MkdirAll(mountPath, os.ModePerm); err != nil {
		return errors.Errorf("Unable to create folder: %s", mountPath)
	}
	if err := runCmd("mount", []string{md.MetadataDrivePath(), mountPath}); err != nil {
		return errors.Errorf("Error: %s", err)
	}

	for key, value := range md.MetadataDriveDatamap() {
		if err := renderData(mountPath, key, value); err != nil {
			return err
		}
	}

	if err := runCmd("umount", []string{mountPath}); err != nil {
		return errors.Errorf("Error: %s", err)
	}
	if err := os.RemoveAll(mountPath); err != nil {
		return errors.Errorf("Unable remove path: %s", mountPath)
	}
	return nil
}

func CreateMetadataDisk(md MetadataDrive) error {
	log.Infoln("Preparing metadrive image...")

	if _, err := os.Stat(md.MetadataDrivePath()); err != nil {
		if err := runCmd("mkfs.msdos", []string{"-C", "-s", "1", md.MetadataDrivePath(), "1440"}); err != nil {
			return errors.Errorf("Error: %s", err)
		}
	} else {
		if err := runCmd("mkfs.msdos", []string{"-s", "1", md.MetadataDrivePath()}); err != nil {
			return errors.Errorf("Error: %s", err)
		}
	}

	return nil
}

