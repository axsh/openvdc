package lxc

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"xi2.org/x/xz"
)

func PrepareCache(cacheFolderPath string, extFolderPath string) error {

	folderState := Exists(cacheFolderPath)
	if folderState == false {
		err := CreateCacheFolder(cacheFolderPath)
		if err != nil {
			return err
		}
	}

	if Exists(filepath.Join(cacheFolderPath, "meta.tar.xz")) == false {
		err := GetFile(cacheFolderPath, extFolderPath, "meta.tar.xz")
		if err != nil {
			errors.Wrapf(err, "Failed downloading file.")
		}
		err = DecompressXz(cacheFolderPath, "meta.tar.xz", cacheFolderPath)
		if err != nil {
			errors.Wrapf(err, "Failed decompressing file.")
		}
	}

	if Exists(filepath.Join(cacheFolderPath, "rootfs.tar.xz")) == false {
		err := GetFile(cacheFolderPath, extFolderPath, "rootfs.tar.xz")
		if err != nil {
			errors.Wrapf(err, "Failed downloading file.")
		}
	}

	return nil
}

func GetFile(cacheFolderPath string, extFolderPath string, fileName string) error {

	filePath := filepath.Join(cacheFolderPath, fileName)

	res, err := http.Get(extFolderPath + fileName)
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

func DecompressXz(cacheFolderPath string, fileName string, outputPath string) error {

	filePath := filepath.Join(cacheFolderPath, fileName)

	f, err := os.Open(filePath)
	if err != nil {
		return errors.Wrapf(err, "Failed reading input file: %s", filePath)
	}

	r, err := xz.NewReader(f, 0)
	if err != nil {
		return errors.Wrapf(err, "Failed creating reader.")
	}

	tr := tar.NewReader(r)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:

			err = os.MkdirAll(hdr.Name, 0777)
			if err != nil {
				log.Fatal(err)
			}
		case tar.TypeReg, tar.TypeRegA:
			w, err := os.Create(filepath.Join(outputPath, hdr.Name))
			if err != nil {
				log.Fatal(err)
			}
			_, err = io.Copy(w, tr)
			if err != nil {
				log.Fatal(err)
			}
			w.Close()
		}
	}

	f.Close()

	return nil
}
