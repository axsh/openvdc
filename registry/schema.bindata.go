// Code generated by go-bindata.
// sources:
// ../schema/v1.json
// DO NOT EDIT!

package registry

import (
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

var _SchemaV1Json = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xd4\x58\x4b\x73\x9b\x3c\x14\xdd\xf3\x2b\x18\x92\xdd\x67\x9b\x38\x93\xcd\xe7\x5d\xa7\x9d\x6e\xd3\x69\x3b\xdd\x64\xa8\x47\x88\x8b\x51\x8a\x24\xaa\x07\xb1\x27\xe3\xff\xde\x01\x83\xcd\x43\xc2\xc4\x75\x9a\x78\x91\x71\xb8\x57\xe7\x70\x74\x75\xf4\xe2\xd9\x71\x5d\x8f\x44\xde\xc2\xf5\x12\xa5\x32\xb9\xf0\x7d\x81\x9e\x66\x2b\xa2\x12\x1d\x6a\x09\x02\x73\xa6\x80\xa9\x19\xe6\xd4\x47\x6b\x99\xf8\x3c\x03\x96\x47\xd8\xa7\x48\x2a\x10\xbe\xc4\x09\x50\xe4\xe7\xf3\xd9\xa3\xe4\xec\xca\x9b\x14\x84\xd7\xbb\x68\xcd\xba\xf0\xfd\x22\x39\xdd\x45\x67\x5c\xac\xfc\x48\xa0\x58\x4d\x6f\xee\x2a\x7c\x85\x8b\x40\x62\x41\x32\x45\x38\x2b\xb0\xf7\x19\xb0\x1f\x9f\x3e\xba\x5f\x41\x72\x2d\x30\xb8\xdf\x81\x66\x29\x52\xe0\x7e\xdb\xf1\x97\x20\xb5\xc9\xa0\x68\xcd\xc3\x47\xc0\x6a\x17\xcb\x04\xcf\x40\x28\x02\xd2\x5b\xb8\x45\x1f\x5d\xd7\x5b\x62\x4e\x29\x30\xb5\x8f\x34\xb0\x52\x09\xc2\x56\x5e\x19\xde\x4e\x76\xed\x15\x51\x29\x8c\x6d\xdc\x56\x3e\x8e\xbf\xea\x8c\xa9\x7d\xa3\x2f\x65\x9c\x33\xb8\x8f\xbd\x85\xfb\x50\x05\xdc\x3d\xa4\x4c\x5f\x0b\x28\xb2\xde\x95\x1f\x41\x4c\x18\x29\x64\x48\x9f\x71\x06\xde\xbe\xd9\x76\xf2\x12\x6c\xba\xc6\xa7\x42\x99\x4e\xd3\x06\xb6\xfa\x2f\x70\xea\xa7\x92\xcd\x13\xf0\x5b\x13\x01\xd1\xbe\x53\x55\xbd\xbb\xc5\x71\x5c\x37\xa8\xbc\xb1\x7f\xc5\x61\x4c\xcb\x2e\x36\xea\xd7\x63\x3d\x54\xb5\xd6\xb1\xaf\xa9\xc1\x23\x36\x9f\x58\xc7\xb2\x53\x9d\xba\x4d\x0b\x07\x4c\xd3\x96\x9e\x83\xf2\x46\x28\x30\xf2\x65\x48\x20\x3a\x3f\x45\x49\x89\xbc\x1d\x8d\x74\x9a\xbf\xb5\x43\x0b\x13\x5c\x64\x71\x73\xda\x36\xb0\xad\xbc\x94\xb0\x65\x8e\x33\x6d\x53\x43\x98\x82\x15\x08\x6f\xd2\x4c\x46\x10\x23\x9d\x16\x3d\x98\x5b\x49\x29\x50\x2e\x36\xcb\x55\x78\x56\xe6\xf3\x4b\x7d\x15\x99\x05\x48\xc4\x08\x77\xc6\xbe\x41\x8b\x84\x40\x9b\x36\x29\x51\x40\xbb\xed\xed\x0b\x62\x95\x35\x3a\xb2\x89\x6c\x05\x83\x0e\xda\xe2\xd1\xf6\x9b\xbb\xf1\xbe\x57\x27\xfd\x16\x46\x57\x56\xb9\x1c\x54\x62\xc0\xd4\x4b\x82\x31\x03\x34\x53\x1b\x73\x2a\x4f\x11\x33\x67\x28\xc2\xf6\x64\x96\x6c\xa4\xd7\x4b\x04\x9d\xc8\xb6\x8b\x2d\x48\x51\x14\x89\x53\x0b\x93\x21\xa5\x40\x94\x9b\xfb\xcf\x87\x9b\xe9\xff\x68\x1a\x7f\x98\x7e\x0e\x9e\x6f\xb7\x87\xa7\x45\xf0\xdf\x75\x57\x5a\x5f\x08\xc9\xf2\xbb\xbf\x51\x12\x73\x41\x91\x2a\x3d\x9e\xe5\x77\xc7\xdf\x17\x0a\x12\xad\xc6\x19\xa2\xcb\xe5\xd8\x9e\xb6\xc6\xe9\x93\xae\xf1\xd2\x70\x36\x70\x8f\x4c\x07\xe3\x06\x79\x90\x8f\x24\xb1\xd2\x1e\xa3\x76\x8f\xce\x16\xfb\xca\x3e\xb2\x48\x86\x92\x97\xf3\x3b\x05\x24\xcd\x55\x3f\x91\x12\x09\x9c\x9c\x97\x6f\x65\x2e\xc8\x18\xbe\xf1\x4e\x69\x8d\x34\x45\xeb\x2f\xcd\xd1\x98\xb7\xb3\x84\x0d\x64\x07\x86\xd1\x8b\xf8\x13\x4b\x39\x8a\x4e\xb1\xc7\xc0\x52\x5c\x30\x13\xa9\x04\x37\xcd\xc3\x7a\x88\x3b\x99\xe0\x5d\xf8\xaf\x92\xfd\xbe\x1d\x9d\x23\x41\xd0\x79\xfb\x7d\x41\x93\xa4\xb3\xa7\x17\xca\x53\xc2\xf4\xda\xe4\x61\xeb\x55\xc7\x3f\x98\xcb\x6f\x2e\xbf\xad\x76\x9d\x25\x74\x50\x46\xa8\xe5\x26\xe4\x6f\x2c\x02\x03\x53\xdc\x78\xb6\xf9\x77\x1a\x22\x08\x09\x62\x6f\xab\x21\x86\x88\x0b\xf4\xb6\x1a\x74\xa8\x99\xea\x1e\xde\x5f\x5b\xc3\xa8\xa3\x06\xa1\xa8\x77\xae\x19\x3c\x67\x0c\xed\x20\x03\x8b\xf0\xe0\x44\xef\xd9\xa6\xda\x89\x96\x5a\xa4\x23\xb8\x7a\x9b\xc5\xe1\x74\xa7\x05\x39\x32\x4b\x12\xc0\xbf\xa4\xa6\x4b\xcb\x81\xff\x45\xba\x6b\xb2\x17\xf3\x58\xf7\x7a\xeb\xbe\xda\xae\x51\x23\xd5\xb8\xea\x3a\xcd\xdf\xfa\x56\x5f\x7e\x9f\xb9\xd4\x6b\x7d\xfb\xe3\x92\xed\x5e\x7f\x21\x17\xe5\x4b\xf8\xfc\xd0\xb6\x90\x53\xfc\x6d\x9d\x3f\x01\x00\x00\xff\xff\xc4\x6c\x33\x2e\x36\x16\x00\x00")

func SchemaV1JsonBytes() ([]byte, error) {
	return bindataRead(
		_SchemaV1Json,
		"../schema/v1.json",
	)
}

func SchemaV1Json() (*asset, error) {
	bytes, err := SchemaV1JsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "../schema/v1.json", size: 5686, mode: os.FileMode(420), modTime: time.Unix(1489657760, 0)}
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
	"../schema/v1.json": SchemaV1Json,
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
	"..": &bintree{nil, map[string]*bintree{
		"schema": &bintree{nil, map[string]*bintree{
			"v1.json": &bintree{SchemaV1Json, map[string]*bintree{}},
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

