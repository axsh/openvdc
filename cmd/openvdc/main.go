package main

import (
	"fmt"
	"os"
	"path"

	"github.com/axsh/openvdc/cmd/openvdc/cmd"
)

// Build time constant variables from -ldflags
var (
	version   string
	sha       string
	builddate string
	goversion string
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
	confPath := path.Join(dir, "config")
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
	cmd.UserConfDir = path.Join(os.Getenv("HOME"), ".openvdc")
	err := setupDefaultUserConfig(cmd.UserConfDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup %s\n", cmd.UserConfDir)
		os.Exit(1)
	}
	cmd.Execute()
}
