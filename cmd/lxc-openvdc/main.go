package main

import (
	"bytes"
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
var containerPath string
var containerName string
var rootfsPath string
var imgPath string

func main() {

	_dist := flag.String("dist", "centos", "Name of the distribution")
	_release := flag.String("release", "7", "Release name/version")
	_arch := flag.String("arch", "amd64", "Container architecture")
	_rootfs := flag.String("rootfs", "", "Rootfs path")
	_containerName := flag.String("name", "", "Container name")
	_containerPath := flag.String("path", "", "Container path")
	_errorLogPath := flag.String("error-log-path", "/etc/openvdc/", "Error log path")
	_imgPath := flag.String("img-path", "127.0.0.1/images", "Image path")
	_cachePath := flag.String("cache-path", "/var/cache/lxc", "Cache path")

	flag.Parse()

	dist := *_dist
	release := *_release
	arch := *_arch
	rootfsPath = *_rootfs
	containerPath = *_containerPath
	containerName = *_containerName
	errorLogPath := *_errorLogPath
	imgPath = *_imgPath
	cachePath := *_cachePath

	lxcPath = "/usr/share/lxc/"
	cacheFolderPath = filepath.Join(cachePath, dist, release, arch)
	imgPath = filepath.Join(imgPath, dist, release, arch)

	var errorLog bytes.Buffer

	err := PrepareCache()
	if err != nil {
		errorLog.Write([]byte(err.Error() + "\n"))
	}

	err = SetupContainerDir()
	if err != nil {
		errorLog.Write([]byte(err.Error() + "\n"))
	}

	err = GenerateConfig()
	if err != nil {
		errorLog.Write([]byte(err.Error() + "\n"))
	}

	if errorLog.Len() > 0 {
		ioutil.WriteFile(filepath.Join(errorLogPath, "lxc-openvdc-ERROR.log"), errorLog.Bytes(), 0644)
	}
}

func SetupContainerDir() error {
	err := os.MkdirAll(rootfsPath, os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "Failed creating container folder.")
	}

	if rootfsPath != "" {
		err = DecompressXz("rootfs.tar.xz", rootfsPath)
		if err != nil {
			return err
		}
	} else {
		return errors.New("RootfsPath not set.")
	}

	return nil
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
			return errors.Wrapf(err, "Failed downloading file.")
		}
		err = DecompressXz("meta.tar.xz", cacheFolderPath)
		if err != nil {
			return errors.Wrapf(err, "Failed decompressing file.")
		}
	}

	if Exists(filepath.Join(cacheFolderPath, "rootfs.tar.xz")) == false {
		err := GetFile("rootfs.tar.xz")
		if err != nil {
			return errors.Wrapf(err, "Failed downloading file.")
		}
	}

	return nil
}

func GenerateConfig() error {
	lxcCfgPath := filepath.Join(lxcPath, "config")
	cfgPath := filepath.Join(containerPath, "config")
	cacheCfgPath := filepath.Join(cacheFolderPath, "config")

	f, err := ioutil.ReadFile(cacheCfgPath)
	if err != nil {
		return errors.Wrapf(err, "Failed reading file: %s.", cacheCfgPath)
	}

	s := "\n# Distribution configuration\n" + string(f[:])
	s = strings.Replace(s, "LXC_TEMPLATE_CONFIG", lxcCfgPath, -1)

	s += fmt.Sprintf(
`
# Container specific configuration
lxc.rootfs = %s
lxc.utsname = %s
 
# Network configuration
lxc.network.type = veth
lxc.network.flags = up
lxc.network.link = virbr0
`, rootfsPath, containerName)

	b := []byte(s)

	err = ioutil.WriteFile(cfgPath, b, 0644)
	if err != nil {
		return errors.Wrapf(err, "Failed writing to file: %s.", cfgPath)
	}

	return nil
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
		errors.Wrapf(err, "Failed unpacking file: %s.", filePath)
	}

	return nil
}
