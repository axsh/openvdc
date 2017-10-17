// Code generated by go-bindata.
// sources:
// ../schema/none.json
// ../schema/v1.json
// ../schema/vm/esxi.json
// ../schema/vm/lxc.json
// ../schema/vm/null.json
// ../schema/vm/qemu.json
// DO NOT EDIT!

package registry

import (
	"github.com/elazarl/go-bindata-assetfs"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _schemaNoneJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x8e\xb1\x4e\x03\x31\x0c\x86\xf7\x3c\x85\x15\x18\x40\x6a\x39\x40\x4c\x59\x41\xac\x48\x08\xb1\xa0\x0a\x85\x9c\xb9\xa6\x6a\xec\xe0\xf8\x86\xaa\xea\xbb\xa3\x5c\xe0\x54\x26\x3a\x64\xc8\xe7\xef\xf7\xef\xbd\x01\xb0\xe7\x25\xac\x31\x79\xeb\xc0\xae\x55\xb3\xeb\xba\x4d\x61\x5a\x36\x7a\xc5\x32\x74\xbd\xf8\x4f\x5d\x5e\xdf\x75\x8d\x9d\xd9\x45\xcd\xf5\x58\x82\xc4\xac\x91\xa9\x66\x9f\x32\xd2\xeb\xc3\x3d\x3c\x63\xe1\x51\x02\xc2\x0b\xa6\xbc\xf5\x8a\x0e\x88\x09\xe1\xe2\x91\x05\x14\x8b\x46\x1a\x80\x69\xbb\xbb\x6c\x6b\x74\x97\xb1\xe6\xf9\x63\x83\x41\x1b\x13\xfc\x1a\xa3\x60\x6f\x1d\xbc\x19\x80\x5f\xcb\x00\xac\xa6\x79\x16\xce\x28\x1a\xb1\x58\x07\xfb\x66\xbc\x07\x4e\x09\x49\x67\x72\xb4\xbb\xa8\x44\x1a\xec\x84\x0f\x0b\x73\x3c\x9b\x5d\xa4\x31\xcd\x7d\x13\xa9\x67\xdb\x9f\xef\xea\x4f\x36\x7b\xf1\xe9\xe6\xd4\xa6\xc9\xbe\xfd\xd7\x36\xf5\x1d\xcc\x77\x00\x00\x00\xff\xff\x57\x3a\x39\x38\x94\x01\x00\x00")

func schemaNoneJsonBytes() ([]byte, error) {
	return bindataRead(
		_schemaNoneJson,
		"schema/none.json",
	)
}

func schemaNoneJson() (*asset, error) {
	bytes, err := schemaNoneJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schema/none.json", size: 404, mode: os.FileMode(420), modTime: time.Unix(1505460640, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _schemaV1Json = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x91\x4d\x4b\x33\x31\x14\x85\xd7\xcd\xaf\x08\x79\xdf\x65\x3b\x51\x70\x35\x5b\xdd\x17\x54\xdc\x48\x91\x69\xe6\xb6\x93\x32\xf9\xe8\xcd\x4d\xad\x0c\xf3\xdf\x25\xf3\x45\xd5\x82\x83\x59\x25\xe7\x9e\xe7\x24\x87\x34\x8c\x73\xa1\x4b\x91\x73\x51\x11\xf9\x90\x4b\x89\xc5\x7b\xb6\xd7\x54\xc5\x6d\x0c\x80\xca\x59\x02\x4b\x99\x72\x46\x16\xe7\x50\x49\xe7\xc1\x9e\x4a\x25\x4d\x11\x08\x50\x06\x55\x81\x29\xe4\xe9\x36\x3b\x04\x67\xff\x89\x65\x0a\xfc\xdf\xab\x63\x6a\x2e\x65\x1a\xae\x7a\x35\x73\xb8\x97\x25\x16\x3b\x5a\xdd\xdc\x0d\xfc\xc0\x95\x10\x14\x6a\x4f\xda\xd9\xc4\xae\x3d\xd8\x97\x87\x7b\xfe\x08\xc1\x45\x54\xc0\x9f\xc1\xf8\xba\x20\xe0\x4f\x7d\x7e\x07\xd1\x87\x87\xe4\x76\xdb\x03\x28\xea\x35\x84\x63\xd4\x08\xa9\xd7\x2b\xe3\x3c\xb9\x34\xd5\xd0\x0d\xd3\x61\xc8\x11\x8c\xf3\x4d\x07\x78\x74\x1e\x90\x34\x04\x91\xf3\xa6\x77\xbd\x29\x67\x0c\x58\x9a\x94\x8b\xcb\x02\xa1\xb6\x7b\xd1\xc9\xed\xf2\xf2\x8a\x99\xe6\xaf\x55\xe7\xe5\x8f\xaf\xbe\xe2\xbf\x28\xdf\xe9\xce\xc2\x7a\x37\xb5\x4f\xab\x99\x76\xe9\x83\x10\xd2\x54\x64\xd2\x3a\x0b\xc3\xd7\x4d\x86\x76\xf9\x1b\x75\x32\xb2\x3e\xab\x2b\xdc\xa2\x61\x8b\xef\x4e\x08\x67\x3d\x5a\x17\xb3\xb2\x6d\xac\xeb\x3f\x3d\xea\x08\x26\xfe\x00\x87\xdd\x86\x8d\xa7\x96\xb5\xec\x33\x00\x00\xff\xff\x56\x01\xb5\x2f\xf9\x02\x00\x00")

func schemaV1JsonBytes() ([]byte, error) {
	return bindataRead(
		_schemaV1Json,
		"schema/v1.json",
	)
}

func schemaV1Json() (*asset, error) {
	bytes, err := schemaV1JsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schema/v1.json", size: 761, mode: os.FileMode(420), modTime: time.Unix(1508232342, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _schemaVmEsxiJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x94\x4f\x8f\x94\x40\x10\xc5\xef\x7c\x8a\x4e\xbb\x37\x07\x59\xcd\x5c\xe4\x66\x34\x26\x9e\x4c\x8c\xf1\xb2\xc1\x49\x2f\x14\x4c\x6f\xa6\xff\x6c\x75\x35\x91\x10\xbe\xbb\x69\x18\x18\x18\x30\x51\x67\x2f\x33\xc3\xaf\x5e\xbd\xaa\x49\xbd\xd0\x46\x8c\xf1\x3b\x97\x1f\x41\x09\x9e\x32\x7e\x24\xb2\x69\x92\x3c\x39\xa3\xe3\x81\xbe\x31\x58\x25\x05\x8a\x92\xe2\xfb\x7d\x32\xb0\x57\x7c\x17\xfa\x0a\x70\x39\x4a\x4b\xd2\xe8\xd0\xfb\xd5\x82\xfe\xf1\xe9\x23\xfb\x06\xce\x78\xcc\x81\x7d\x07\x65\x4f\x82\x20\x65\xb5\x4a\xc0\xfd\x92\x43\x1b\x35\x16\x82\xde\x3c\x3e\x41\x4e\x03\x43\x78\xf6\x12\xa1\xe0\x29\x7b\x88\x18\x1b\x55\x11\x63\x59\x5f\xb7\x68\x2c\x20\x49\x70\x3c\x65\xed\xa0\x38\xe4\x46\x29\xd0\x34\x91\x99\xb7\x23\x94\xba\xe2\x3d\xee\x76\xd1\xbc\x36\x69\x41\x7b\x35\xcd\xeb\xc9\xb8\xe6\x99\x64\x8b\x76\x25\xf5\xa1\xce\xad\xdf\x1a\x27\x35\x41\x05\xc8\x77\x63\xa1\x80\x52\xf8\x53\x58\xed\xed\xca\x44\x81\x32\xd8\x1c\xaa\xc7\x9b\x9c\x6e\x5f\xe5\x45\xd6\xd0\xa6\x80\x43\x85\xc6\x5b\xb7\xe5\x23\x10\x45\x73\x71\xf1\x5a\x3e\x7b\xf8\x42\xa0\x82\x9a\xd0\xc3\x54\x92\x67\x78\x39\x47\x7b\x7d\xcd\x6e\xf3\x2e\x61\x55\x2c\x45\x0e\x7f\xb3\xc0\x38\xa5\xbd\x1c\x7d\x23\x8e\xe7\xca\x2a\x94\xf3\x8e\x09\x64\xb3\x8e\x8d\x98\x2e\xa7\xcc\xd9\x3a\xae\xbb\x65\x75\x15\xd0\x31\xa6\x40\xc7\x2b\x6d\x7f\x09\x0d\x6b\x0a\xca\x52\xb3\xc6\xf5\x49\xe8\x35\x55\x22\xdf\x2e\xd8\x63\xe3\xf8\x02\x66\xb3\xa7\x6e\xae\x0f\x26\xa2\x28\xf0\x5f\xff\xac\x15\x44\x80\xfd\x8b\xe4\xe7\xc3\x7d\xfc\x5e\xc4\xe5\x87\xf8\x73\xd6\xbe\xeb\x2e\x4f\x69\xf6\xfa\x8e\xff\x71\xb0\xb4\xf5\xfe\x7f\x26\x97\x06\x95\xa0\x3e\xf5\xb6\xde\x2f\xfc\xa3\xeb\x5f\xc3\x77\xf8\xec\xa2\x2e\xfa\x1d\x00\x00\xff\xff\xdc\x74\x08\x75\x3e\x05\x00\x00")

func schemaVmEsxiJsonBytes() ([]byte, error) {
	return bindataRead(
		_schemaVmEsxiJson,
		"schema/vm/esxi.json",
	)
}

func schemaVmEsxiJson() (*asset, error) {
	bytes, err := schemaVmEsxiJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schema/vm/esxi.json", size: 1342, mode: os.FileMode(420), modTime: time.Unix(1508232342, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _schemaVmLxcJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x56\xcd\x6e\xdb\x3c\x10\xbc\xeb\x29\x08\x25\xb7\xcf\xfe\x94\x14\xb9\xd4\xb7\xa2\x45\x81\x9e\x5a\x14\x45\x2f\x81\x2a\x50\xd4\x4a\x66\x2a\xfe\x64\x49\xba\x36\x0c\xbd\x7b\x41\x5b\xb6\xfe\x13\xa7\x4e\x82\x1c\x0c\x58\xb3\xdc\xd9\x25\x77\x34\xd4\x36\x20\x24\xbc\x34\x6c\x09\x82\x86\x0b\x12\x2e\xad\xd5\x8b\x28\xba\x33\x4a\xce\xf7\xe8\xff\x0a\x8b\x28\x43\x9a\xdb\xf9\xd5\x4d\xb4\xc7\x2e\xc2\x99\xcf\xcb\xc0\x30\xe4\xda\x72\x25\x7d\xee\x57\x0d\xf2\xe7\xa7\x8f\xe4\x3b\x18\xe5\x90\x01\xf9\x01\x42\x97\xd4\xc2\x82\xac\x44\x54\xae\xd9\x3e\xcb\x6e\x34\xf8\xe5\x2a\xbd\x03\x66\xf7\x18\xc2\xbd\xe3\x08\x59\xb8\x20\xb7\x01\x21\x87\x55\x01\x21\xf1\x2e\xae\x51\x69\x40\xcb\xc1\x84\x0b\xb2\xdd\xaf\x48\x98\x12\x02\xa4\x3d\x22\x2d\x6e\x63\x91\xcb\x22\xdc\xc1\xd5\x2c\x68\xc7\x8e\x6b\x41\x3a\x71\xac\xb7\x43\xea\x2e\x6b\x20\xee\x64\x0b\x2e\x93\x15\xd3\x6e\xac\x1a\x97\x16\x0a\xc0\x70\x76\x08\x64\x90\x53\x57\xfa\xce\xae\x07\x24\x02\x84\xc2\x4d\x52\xa4\x67\x31\x9d\xdf\xca\xb3\xb4\x21\x55\x06\x49\x81\xca\x69\x33\xc6\x43\x11\xe9\xa6\x61\x71\x92\xdf\x3b\xf8\x62\x41\xf8\xd5\x16\x1d\x1c\x43\xbc\x06\x9b\x69\x6c\xfb\xc3\xac\x46\xe7\xe2\x5b\xc5\x9c\x32\x38\xa5\x81\x43\x95\x6d\x33\xf3\x11\x35\xd6\x91\x81\x26\xdb\x19\x47\x20\x6e\x65\x8c\xa8\xb4\x5b\xa5\x8d\x0d\xd5\x3a\xeb\x46\x07\xfa\x3c\xa8\x14\xec\xb2\xb7\x76\x37\x09\x09\x43\x14\x84\xb6\x9b\x21\xbc\x2a\xa9\x1c\xa2\x82\xb2\xf1\x80\x5e\x6e\x4c\xd8\x01\xe3\xd6\x53\xd5\x5e\xef\x49\x68\x96\xe1\x53\x37\xab\xa9\xb5\x80\x3b\x1b\xf9\x75\x7b\x35\x7f\x4f\xe7\xf9\x87\xf9\xe7\x78\xfb\xae\x6a\x9e\x16\xf1\x7f\x97\xe1\x64\x61\xae\x57\x37\xff\x52\x39\x57\x28\xa8\xdd\xa9\x5e\xaf\x6e\x3a\xfc\x41\xff\x5f\xd5\x91\x5e\xb9\x66\x89\xad\x2d\x6e\x4c\x7c\x3d\x49\xf9\x97\x88\x4b\xee\xdd\xb2\xa7\xc1\x94\x1a\x3e\x4a\xf5\x10\x1d\x79\x50\x71\xe3\xfe\x38\x75\x24\x9d\x70\xd5\x3b\x20\x84\x12\xa8\x19\xca\xf7\x89\x34\x14\xd9\xf2\x7c\x8e\x62\xb8\xd1\xc7\x38\x1e\x9e\xe8\x71\x3a\x82\xae\xbf\xb5\x4f\xf3\xba\x89\x70\x39\x11\x99\x38\xfe\x30\x53\x7f\x64\xa9\x68\xf6\x94\x51\x4e\xd8\x8d\x67\xe3\xc6\xa2\xea\xeb\xf6\x30\x96\x16\x1a\xbf\xba\x36\xea\xd6\xde\x86\xc2\x56\x14\x39\x3d\x7f\x4f\x6f\x4c\xa8\xad\xfb\xc5\x77\x56\x72\xe9\xd6\x7d\x5d\x5d\x22\xe4\x9e\xf6\x22\x6a\x86\x1e\xb5\xed\x29\x6a\x79\x4f\xd4\xb3\x9b\xd1\x52\xa9\x33\x9b\x54\xbd\x42\x21\x06\xd2\xaa\xc1\x7d\xf9\xfc\x75\x32\x48\x39\x95\x2f\x5f\x27\x87\x4c\x21\x7d\xf9\x3a\x2e\x75\xd2\xba\x17\xaa\x33\x79\xdd\x71\x41\x8b\x93\xee\xba\x29\x67\x9c\x30\x9e\xc9\x17\xa3\x33\xc2\xda\x55\x13\x87\xe5\x23\xf9\x1d\x23\x6c\x6e\x78\x87\x7c\x42\x85\x4b\x60\xbf\x8d\x13\xc9\xc8\x87\xda\x49\xbd\x1d\x08\x4e\xce\x1d\xdc\x3f\xa3\xfe\xdf\xdd\x73\xf7\x03\x38\xf0\xbf\x2a\xf8\x1b\x00\x00\xff\xff\xe7\x65\x42\x4c\xc2\x0d\x00\x00")

func schemaVmLxcJsonBytes() ([]byte, error) {
	return bindataRead(
		_schemaVmLxcJson,
		"schema/vm/lxc.json",
	)
}

func schemaVmLxcJson() (*asset, error) {
	bytes, err := schemaVmLxcJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schema/vm/lxc.json", size: 3522, mode: os.FileMode(420), modTime: time.Unix(1505460640, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _schemaVmNullJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x92\x41\x8b\xdb\x40\x0c\x85\xef\xfe\x15\x62\xda\x43\x0b\x4e\xdd\x42\x4f\xbe\xb6\x14\x7a\x2a\x94\x65\x2f\x4b\x30\x93\xb1\xe2\x4c\xf0\x8c\x26\x1a\x4d\xc0\x84\xfc\xf7\xc5\x71\xe2\x38\xbb\xde\xbd\x64\x0f\x3e\xf8\x7b\x7a\x4f\xb2\xa5\x43\x06\xa0\x3e\x47\xb3\x41\xa7\x55\x09\x6a\x23\x12\xca\xa2\xd8\x46\xf2\x8b\x81\x7e\x23\x6e\x8a\x9a\xf5\x5a\x16\xdf\x7f\x16\x03\xfb\xa4\xf2\xde\x57\x63\x34\x6c\x83\x58\xf2\xbd\xf7\x5f\x40\xff\xf8\xfb\x17\xfc\xc7\x48\x89\x0d\xc2\x03\xba\xd0\x6a\xc1\x12\xf6\xae\xf0\xa9\x6d\xe1\xcb\x1f\x62\x10\x8c\x62\x7d\x03\xe4\xdb\xee\xeb\x90\x24\x5d\xc0\x3e\x82\x56\x5b\x34\x32\x30\xc6\x5d\xb2\x8c\xb5\x2a\xe1\x29\x03\xb8\x54\x65\x00\xcb\x93\x1e\x98\x02\xb2\x58\x8c\xaa\x84\xc3\x50\x51\x19\x72\x0e\xbd\x8c\x64\x92\x1d\x85\xad\x6f\xd4\x09\x1f\xf3\x6c\xaa\x8d\xb5\xe8\x93\x1b\xfb\x9d\xc8\x79\x72\x75\x26\xcb\x1b\xbb\xb3\xbe\xda\x9b\x90\xe6\xda\x59\x2f\xd8\x20\xab\xfc\x22\xd4\xb8\xd6\xa9\xed\x47\xfb\xf1\x2a\xc4\xa1\x23\xee\xaa\x66\x75\x57\xd2\xfd\xa3\x7c\xc8\x18\x9e\x6a\xac\x1a\xa6\x14\xe2\x5c\x8e\x66\xd6\xdd\x35\x25\x79\xbb\x4b\xf8\x57\xd0\xf5\xd5\xc2\x09\x47\xc9\x9e\xe1\x75\x1d\x87\x97\xdb\x3c\xce\xee\xc5\xb0\x8e\x9b\x2a\x8a\x6e\xf0\x9d\x4b\xc8\xdf\xde\xba\x27\x8f\xa3\x0e\xbd\x43\xb3\xdc\x02\x0a\xd3\x77\xc3\xa8\xe5\xc6\x52\x63\x14\xa6\x6e\x8a\x18\x57\x44\x32\x9e\xd2\xcc\x8f\x1c\x1a\x0f\xdf\x92\xf5\xcf\x31\x7b\x0e\x00\x00\xff\xff\xaa\xf3\x11\xf0\xa2\x03\x00\x00")

func schemaVmNullJsonBytes() ([]byte, error) {
	return bindataRead(
		_schemaVmNullJson,
		"schema/vm/null.json",
	)
}

func schemaVmNullJson() (*asset, error) {
	bytes, err := schemaVmNullJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schema/vm/null.json", size: 930, mode: os.FileMode(420), modTime: time.Unix(1505460640, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _schemaVmQemuJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x54\xcd\x8e\x9b\x3c\x14\xdd\xf3\x14\x96\xbf\xd9\x7d\x49\x33\x1d\x65\x53\x76\x55\xab\x4a\x5d\x55\xaa\xaa\x6e\x46\x14\x39\xe6\x42\x3c\x83\x7f\x72\x6d\x13\x45\x11\xef\x5e\x41\x02\x31\xc1\xa3\xa6\x33\xdd\xc1\xb9\x7f\xe7\x72\x0e\xf7\x98\x10\x42\xef\x2c\xdf\x82\x64\x34\x25\x74\xeb\x9c\x49\x57\xab\x27\xab\xd5\xf2\x84\xbe\xd3\x58\xad\x0a\x64\xa5\x5b\xde\xaf\x57\x27\xec\x3f\xba\xe8\xea\x0a\xb0\x1c\x85\x71\x42\xab\xae\xf6\x9b\x01\xf5\xf3\xf3\x27\xf2\x1d\xac\xf6\xc8\x81\xfc\x00\x69\x6a\xe6\x20\x25\x8d\x5c\xed\x40\xfa\x53\x99\x3b\x18\xe8\xf2\xf5\xe6\x09\xb8\x3b\x61\x08\x3b\x2f\x10\x0a\x9a\x92\xc7\x84\x90\x21\x2b\x21\x24\xeb\xe3\x06\xb5\x01\x74\x02\x2c\x4d\xc9\xf1\x94\x91\x73\x2d\x25\x28\x37\x22\x41\x6f\xeb\x50\xa8\x8a\xf6\x70\xbb\x48\xc2\xd8\x98\x0b\xca\xcb\x71\x5e\x8f\x0c\x34\xcf\x48\x36\x29\x97\x42\xe5\x0d\x37\x3e\x36\x4e\x28\x07\x15\x20\x5d\x0c\x81\x02\x4a\xe6\xeb\x8e\xda\xfb\x59\x13\x09\x52\xe3\x21\xaf\x36\x6f\xea\xf4\x76\x2a\xff\x84\x86\xd2\x05\xe4\x15\x6a\x6f\x6c\xac\x0f\x43\x64\x87\x4b\x17\xaf\xc4\xce\xc3\x57\x07\xb2\xcb\x76\xe8\x61\x0c\x89\x33\x78\x91\xe3\x78\xad\x66\x1b\xd5\xa5\xa3\x8a\x25\xe3\x70\x0b\x81\x61\xca\xf1\x22\x7a\xc4\x8e\xe7\xc8\xcc\x94\x61\xc5\x08\x64\x41\x45\xc4\xa6\xd3\x29\x21\x36\xb7\xeb\x62\x1a\x9d\x19\x74\xb0\x29\xb8\xed\x55\x6e\xaf\x84\x82\x39\x0a\xd2\xb8\xc3\x1c\x6e\x6a\xa6\xe6\xa8\x64\x3c\x1e\x30\xdb\x83\xa5\x13\x30\x0b\xde\xda\x30\xbf\x6b\xc2\x8a\x02\xff\x76\x59\xc3\x9c\x03\xec\x0f\xc9\xaf\xc7\xfb\xe5\x07\xb6\x2c\x3f\x2e\xbf\x64\xc7\x87\xf6\xf2\x96\x66\xff\xdf\xd1\x17\x07\x0b\xd3\xac\x5f\x33\xb9\xd4\x28\x99\xeb\x5d\x6f\x9a\xf5\xa4\x7f\x72\xfd\xd4\x4e\xac\xd7\x1d\x8b\x5c\x48\x56\x41\xcc\x7a\x57\x86\x7a\xc9\x1c\xb1\x3b\x16\x63\x9d\x44\xb6\xa6\x85\xde\xab\x5a\xb3\x22\xf7\x58\xff\xa1\x7e\xf2\xad\x2e\x3b\x7b\x14\xf1\xd6\x7c\x0b\xfc\xd9\x7a\x99\x47\xac\x7b\x13\xb7\xa1\xc1\x6b\x6a\x47\x7e\x37\x6f\x14\xfd\x59\x28\xb2\xfd\xb5\xdc\x3b\xae\xf7\x0f\xa1\xca\xd9\x5c\xe5\x51\xb3\xe8\x09\x98\x7e\xf5\x39\xeb\xe8\x99\xf2\x16\xf2\xe7\x46\xc6\x8c\xb2\xd1\xba\x86\xe0\xb7\x0b\x8e\x6d\xc9\x6a\x0b\xc9\xc0\xad\x4d\xda\xe4\x77\x00\x00\x00\xff\xff\x0e\xae\x27\x8a\xb2\x07\x00\x00")

func schemaVmQemuJsonBytes() ([]byte, error) {
	return bindataRead(
		_schemaVmQemuJson,
		"schema/vm/qemu.json",
	)
}

func schemaVmQemuJson() (*asset, error) {
	bytes, err := schemaVmQemuJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "schema/vm/qemu.json", size: 1970, mode: os.FileMode(420), modTime: time.Unix(1505460640, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"schema/none.json": schemaNoneJson,
	"schema/v1.json": schemaV1Json,
	"schema/vm/esxi.json": schemaVmEsxiJson,
	"schema/vm/lxc.json": schemaVmLxcJson,
	"schema/vm/null.json": schemaVmNullJson,
	"schema/vm/qemu.json": schemaVmQemuJson,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"schema": &bintree{nil, map[string]*bintree{
		"none.json": &bintree{schemaNoneJson, map[string]*bintree{}},
		"v1.json": &bintree{schemaV1Json, map[string]*bintree{}},
		"vm": &bintree{nil, map[string]*bintree{
			"esxi.json": &bintree{schemaVmEsxiJson, map[string]*bintree{}},
			"lxc.json": &bintree{schemaVmLxcJson, map[string]*bintree{}},
			"null.json": &bintree{schemaVmNullJson, map[string]*bintree{}},
			"qemu.json": &bintree{schemaVmQemuJson, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}


func assetFS() *assetfs.AssetFS {
	assetInfo := func(path string) (os.FileInfo, error) {
		return os.Stat(path)
	}
	for k := range _bintree.Children {
		return &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: assetInfo, Prefix: k}
	}
	panic("unreachable")
}
