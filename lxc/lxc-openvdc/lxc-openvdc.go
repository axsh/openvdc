package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
	"github.com/pkg/errors"
)

var lxcPath string
var cacheFolderPath string
var imgPath string
var containerName string
var containerPath string
var rootfsPath string

func main() {

	//TODO: Set these depending on passed args
	cacheFolderPath = "/var/cache/lxc/centos/7/amd64/"
	containerName = "test"
	lxcPath = "/var/lib/lxc/"
	imgPath = "127.0.0.1/images/centos/7/amd64/"
	containerPath = filepath.Join(lxcPath, containerName)
	rootfsPath = filepath.Join(containerPath, "rootfs")

	err := PrepareCache()
	if err != nil {
		fmt.Println(err)
	}

	SetupContainerDir()

	GenerateConfig()
}

func SetupContainerDir() {
	os.MkdirAll(rootfsPath, os.ModePerm)

	DecompressXz("rootfs.tar.xz", rootfsPath)
}

func PrepareCache() error {

	if Exists(cacheFolderPath) == false {
		err := CreateCacheFolder(cacheFolderPath)
		if err != nil {
			return err
		}
	}

	if Exists(filepath.Join(cacheFolderPath, "meta.tar.xz")) == false {
		err := GetFile("meta.tar.xz")
		if err != nil {
			errors.Wrapf(err, "Failed downloading file.")
		}
		err = DecompressXz("meta.tar.xz", cacheFolderPath)
		if err != nil {
			errors.Wrapf(err, "Failed decompressing file.")
		}
	}

	if Exists(filepath.Join(cacheFolderPath, "rootfs.tar.xz")) == false {
		err := GetFile("rootfs.tar.xz")
		if err != nil {
			errors.Wrapf(err, "Failed downloading file.")
		}
	}

	return nil
}

func GenerateConfig() {

	lxcCfgPath := filepath.Join(lxcPath, "config")
	cfgPath := filepath.Join(containerPath, "config")

	if Exists(cfgPath) == false {

		f, err := ioutil.ReadFile(filepath.Join(cacheFolderPath, "config-user"))
		if err != nil {
			fmt.Print(err)
		}

		s := string(f[:])
		s = strings.Replace(s, "LXC_TEMPLATE_CONFIG", lxcCfgPath, -1)
		b := []byte(s)

		add := "\n\nlxc.rootfs = " + rootfsPath +
			"\nlxc.utsname = " + containerName

		res := append(b, add...)

		err = ioutil.WriteFile(cfgPath, res, 0644)
	}
}

func GetFile(fileName string) error {

	filePath := filepath.Join(cacheFolderPath, fileName)

	res, err := http.Get(imgPath + "/" + fileName)
	if err != nil {
		return errors.Wrapf(err, "Failed Http.Get for file: %s", fileName)
	}

	defer res.Body.Close()

	fileContent, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrapf(err, "Failed reading http response for file: %s", fileName)

	}

	f, err := os.Create(filePath)

	if err != nil {
		return errors.Wrapf(err, "Failed creating file: %s", fileName)
	}
	defer f.Close()

	_, err = f.Write(fileContent)
	if err != nil {
		return errors.Wrapf(err, "Failed writing to file: %s", fileName)
	}

	return nil
}

func Exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func CreateCacheFolder(folderPath string) error {
	err := os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "Failed to create cache folder: %s", folderPath)
	} else {
		return nil
	}
}

func DecompressXz(fileName string, outputPath string) error {

	filePath := filepath.Join(cacheFolderPath, fileName)

	err := archiver.TarXZ.Open(filePath, outputPath)

	if err != nil {
		return err
	}

	return nil
}
