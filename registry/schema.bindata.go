// Code generated by go-bindata.
// sources:
// schema/v1.json
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

var _schemaV1Json = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xd4\x55\x4d\x8f\xda\x30\x10\xbd\xe7\x57\x58\xde\x3d\x66\x71\x57\xea\x29\xd7\xf6\xbe\x52\x5b\xf5\x52\xa1\xc8\x38\x43\xe2\x6d\xfc\xd1\xb1\xcd\xb2\xaa\xf8\xef\x95\xf3\x45\x02\x04\x41\xc5\x1e\xb8\xc1\xcc\xbc\x97\x37\x33\x2f\x93\xbf\x09\x21\x54\x16\x34\x23\xb4\xf2\xde\xba\x8c\x31\xe4\x6f\x8b\x52\xfa\x2a\xac\x82\x03\x14\x46\x7b\xd0\x7e\x21\x8c\x62\x7c\xeb\x2a\x66\x2c\xe8\x4d\x21\x98\xe2\xce\x03\x32\x27\x2a\x50\x9c\x6d\x9e\x17\xaf\xce\xe8\x07\x9a\x46\xc2\xc7\x36\xda\xb3\x66\x8c\xc5\xe4\x53\x1b\x5d\x18\x2c\x59\x81\x7c\xed\x9f\x3e\x7d\xee\xf0\x1d\xae\x00\x27\x50\x5a\x2f\x8d\x8e\xd8\x17\x0b\xfa\xe7\xd7\x2f\xe4\x1b\x38\x13\x50\x00\xf9\x01\xca\xd6\xdc\x03\xf9\xde\xf2\x37\x20\xff\x6e\x21\x56\x9b\xd5\x2b\x08\xdf\xc6\x2c\x1a\x0b\xe8\x25\x38\x9a\x91\xd8\x23\x21\x34\x17\x46\x29\xd0\x7e\x88\x8c\xb0\xce\xa3\xd4\x25\x6d\xc2\xbb\xb4\xad\xf7\xd2\xd7\x70\x69\xf1\x54\xf9\x65\xfc\x5d\x33\xa7\xea\x47\xbd\x34\x71\xa3\xe1\x65\x4d\x33\xf2\xab\x0b\x90\x01\xd2\xa4\x1f\x11\x62\x96\x3e\xb0\x9e\xd4\x31\x6d\x34\xd0\xa1\x68\x97\x5e\x8e\xac\xb7\xe2\xff\x80\x3a\xd4\xf5\x08\xd9\xfd\x5a\x26\xfd\xbf\x86\x8b\x22\xfc\x09\x12\xa1\x18\xda\xe9\x26\x7d\x38\x96\x84\x90\x65\xbb\xe0\xfe\x01\xfb\x5d\x36\xcd\x8d\xe6\x76\xc4\xb9\x9f\x66\xaf\x62\x98\xe5\x09\x6f\xcc\xf9\x63\x76\x87\x07\x93\xe9\x6b\x26\x38\xd0\x41\x4d\xf4\xec\x95\x8f\x42\xcb\xa3\x79\xed\x26\x2e\x89\xcb\xb8\xcb\x46\x27\x2e\x9a\xf4\x39\xa2\x53\x52\xe7\x1b\x61\xc3\x9c\x14\xa9\x3d\x94\x80\x34\x1d\x27\x0b\x58\xf3\x50\x47\xf9\xcf\xb3\xa4\x0a\x94\xc1\xf7\xbc\x5c\xdd\x94\xf9\xf6\x52\x3f\x44\x66\xbd\x15\xb9\x54\xbc\x3c\xda\xd4\xcc\x79\x21\x67\xdc\x42\xce\x38\x86\x9c\x73\xcd\x81\xa8\x56\xb8\x79\xd3\xb5\xe1\x45\x1e\xb0\xbe\x80\x2b\x3d\xcc\xaf\x0d\x2a\x1e\x55\xd0\x80\xf2\xfc\xa3\x44\x05\xe2\xb7\x0b\x2a\x3f\xe1\xd8\xab\x75\xf7\x64\x57\xf3\x24\x33\x9c\xa7\xdf\xe3\xe3\x19\x5d\x71\x28\x9a\xe3\x7b\x97\x97\x62\xfa\xd9\xb8\xe7\x53\x31\x5d\x4c\xfc\xe2\x25\xbb\xe4\x5f\x00\x00\x00\xff\xff\x1b\x02\xb9\x63\x63\x09\x00\x00")

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

	info := bindataFileInfo{name: "schema/v1.json", size: 2403, mode: os.FileMode(420), modTime: time.Unix(1480820698, 0)}
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
	"schema/v1.json": schemaV1Json,
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
		"v1.json": &bintree{schemaV1Json, map[string]*bintree{}},
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

