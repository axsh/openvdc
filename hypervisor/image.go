package hypervisor

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"xi2.org/x/xz"
)

func PrepareCache(cacheFolderPath string, img string) error {

	folderState := CacheFolderExists(cacheFolderPath)
	if folderState == false {
		err := createCacheFolder(cacheFolderPath)
		if err != nil {
			return err
		}
		FetchImage(imgPath)
		DecompressXz(imgPath, cacheFolderPath)
	}

	return nil
}

func FetchImage(imgPath) error {

	//TODO: Fetch image from local img server.

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

func DecompressXz(inputFilePath string, outputFilePath string) error {

	f, err := os.Open(inputFilePath)
	if err != nil {
		return errors.Wrapf(err, "Failed reading input file: %s", inputFilePath)
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
			w, err := os.Create(outputFilePath + hdr.Name)
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
