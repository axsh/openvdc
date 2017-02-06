package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/axsh/openvdc/cmd/openvdc/cmd"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
)

const defaultTomlConfig = `
[api]
# endpoint = 127.0.0.1:5000
[mesos]
# master = 127.0.0.1:5050
`

func setupDefaultUserConfig(dir string) error {
	stat, err := os.Stat(dir)
	if err == nil {
		if stat.IsDir() {
			// nothing to do
			return nil
		} else {
			return fmt.Errorf("%s is not directory", dir)
		}
	}
	err = os.Mkdir(dir, 0755)
	if err != nil {
		return err
	}
	confPath := filepath.Join(dir, "config.toml")
	// Install default configuration file.
	return ioutil.WriteFile(confPath, []byte(defaultTomlConfig), 0644)
}

func main() {
	err := setupDefaultUserConfig(util.UserConfDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup %s: %v\n", util.UserConfDir, err)
		os.Exit(1)
	}
	cmd.Execute()
}
