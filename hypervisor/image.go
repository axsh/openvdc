package hypervisor

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"xi2.org/x/xz"
)

func PrepareCache(cacheFolderPath string, extFolderPath string) error {

	folderState := CacheFolderExists(cacheFolderPath)
	if folderState == false {
		err := CreateCacheFolder(cacheFolderPath)
		if err != nil {
			return err
		}
		GetFile(cacheFolderPath, extFolderPath, "meta.tar.xz")
		DecompressXz(cacheFolderPath, "meta.tar.xz")
	}

	return nil
}

func GetFile(cacheFolderPath string, extFolderPath string, fileName string) error {

	filePath := filepath.Join(cacheFolderPath, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	res, err := http.Get(filepath.Join(extFolderPath, fileName))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, err = io.Copy(f, res.Body)
	if err != nil {
		return err
	}

	return nil
}

func CacheFolderExists(folderPath string) bool {
	if _, err := os.Stat(folderPath); err != nil {
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

func DecompressXz(cacheFolderPath string, fileName string) error {

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
			w, err := os.Create(filepath.Join(cacheFolderPath, hdr.Name))
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
