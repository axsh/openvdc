package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/axsh/openvdc/cmd/openvdc/cmd"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
)

func setupDefaultUserConfig(dir string) error {
	stat, err := os.Stat(dir)
	if os.IsExist(err) && !stat.IsDir() {
		return fmt.Errorf("")
	} else if os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
	}
	// Install default configuration file here.
	confPath := filepath.Join(dir, "config")
	_, err = os.Open(confPath)
	if os.IsNotExist(err) {
		f, err := os.Create(confPath)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return nil
}

func main() {
	err := setupDefaultUserConfig(util.UserConfDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup %s\n", util.UserConfDir)
		os.Exit(1)
	}
	cmd.Execute()
}
