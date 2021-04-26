package asset

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

var _res_services_taskservice_conf_ini = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x1c\xcb\x41\x0a\xc2\x40\x0c\x46\xe1\x7d\x4e\xf1\x43\x0f\x30\x2d\xba\x69\x61\xae\x20\xee\xc5\xc5\xa0\x71\x46\x8c\xa6\x24\x69\xd1\xdb\x4b\xbb\x7b\x8b\xef\x5d\xee\xfc\x28\x8b\xc4\x95\x28\x8a\xbf\x4e\xcb\x1b\x19\x43\xdf\x93\x3c\x3d\xf8\x83\x8c\xe9\xd8\x8f\x23\x89\x56\x67\x5b\xd9\x90\xd1\x22\xe6\x29\x25\xd1\x5b\x91\xa6\x1e\xbb\x48\xdb\x9e\xd6\x21\x89\x56\x0a\xfb\x21\xe3\x40\x9d\x68\x3d\x97\x68\xc8\xa0\x4e\xa3\xb1\xf9\x9e\xb3\xe9\x77\x13\xf4\x0f\x00\x00\xff\xff\x43\xff\x5f\x8b\x7f\x00\x00\x00")

func res_services_taskservice_conf_ini() ([]byte, error) {
	return bindata_read(
		_res_services_taskservice_conf_ini,
		"Res/services/TaskService/conf.ini",
	)
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
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
var _bindata = map[string]func() ([]byte, error){
	"Res/services/TaskService/conf.ini": res_services_taskservice_conf_ini,
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
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() ([]byte, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"Res": &_bintree_t{nil, map[string]*_bintree_t{
		"services": &_bintree_t{nil, map[string]*_bintree_t{
			"TaskService": &_bintree_t{nil, map[string]*_bintree_t{
				"conf.ini": &_bintree_t{res_services_taskservice_conf_ini, map[string]*_bintree_t{
				}},
			}},
		}},
	}},
}}
