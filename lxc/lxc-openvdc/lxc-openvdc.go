package main

import (
	"flag"
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

	_dist := flag.String("dist", "centos", "Name of the distribution")
	_release := flag.String("release", "7", "Release name/version")
	_arch := flag.String("arch", "amd64", "Container architecture")
	_containerName := flag.String("name", "", "Container name")
	_rootfs := flag.String("rootfs", "", "Rootfs path")
	_path := flag.String("path", "", "Container path")

	flag.Parse()

	dist := *_dist
	release := *_release
	arch := *_arch
	containerName := *_containerName

	lxcPath = "/usr/share/lxc/"
	cacheFolderPath = filepath.Join("/var/cache/lxc", dist, release, arch)
	imgPath = filepath.Join("127.0.0.1/images", dist, release, arch)

	containerPath = *_path

	rootfsPath = *_rootfs

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

	f, err := ioutil.ReadFile(filepath.Join(cacheFolderPath, "config"))
	if err != nil {
		fmt.Print(err)
	}

	s := string(f[:])
	s = strings.Replace(s, "LXC_TEMPLATE_CONFIG", lxcCfgPath, -1)
	b := []byte(s)

	err = ioutil.WriteFile(cfgPath, b, 0644)
}

func GetFile(fileName string) error {

	filePath := filepath.Join(cacheFolderPath, fileName)

	res, err := http.Get("http://" + imgPath + "/" + fileName)
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
