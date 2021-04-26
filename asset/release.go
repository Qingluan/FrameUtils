package asset

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func AssetAsFile(name string) (string, error) {
	buf, err := Asset(name)
	if err != nil {
		return "", err
	}
	temp := os.TempDir()
	dir := filepath.Dir(name)
	if err := os.MkdirAll(filepath.Join(temp, dir), os.ModePerm); err == nil {
		ioutil.WriteFile(filepath.Join(temp, name), buf, os.ModePerm)
		return filepath.Join(temp, name), nil
	} else {
		return "", err
	}
}
