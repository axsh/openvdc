package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/axsh/openvdc/cmd/openvdc/cmd"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	homedir "github.com/mitchellh/go-homedir"
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
	//http://stackoverflow.com/questions/7922270/obtain-users-home-directory
	path, err := homedir.Dir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get $HOME")
		os.Exit(1)
	}
	util.UserConfDir = filepath.Join(path, ".openvdc")
	err = setupDefaultUserConfig(util.UserConfDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup %s\n", util.UserConfDir)
		os.Exit(1)
	}
	cmd.Execute()
}
