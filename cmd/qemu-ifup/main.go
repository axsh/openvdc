package main

import (
	"os"
	"os/exec"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

var DefaultConfPath string

// todo: error handling

func runCmd(cmd string, args []string) {
	c := exec.Command(cmd, args...)
	c.Run()
	fmt.Println(c.Args)
}

func init () {
	viper.SetDefault("bridges.type", "linux")
	viper.SetDefault("bridges.name", "br0")
	viper.SetConfigFile(DefaultConfPath)
	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		if viper.ConfigFileUsed() == DefaultConfPath && os.IsNotExist(err) {
			// Ignore default conf file does not exist.
			return
		}
		log.WithError(err).Fatalf("Failed to load config %s", viper.ConfigFileUsed())
	}
}

func main () {
	config := viper.GetViper()
	ifname := os.Args[1]
	switch config.GetString("bridges.type") {
	case "linux":
		runCmd("brctl", []string{"addif", config.GetString("bridges.name"), ifname})
	case "ovs":
		runCmd("ovs-vsctl", []string{"add-port", config.GetString("bridges.name"), ifname})
	default:
		// throw error
	}
	runCmd("ip", []string{"link", "set", "dev", ifname, "up"})

	fmt.Println(config)
	fmt.Println(os.Args)
}
