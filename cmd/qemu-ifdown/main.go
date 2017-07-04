package main

import (
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

var DefaultConfPath string

func runCmd(cmd string, args []string) {
	c := exec.Command(cmd, args...)
	if err := c.Run(); err != nil {
		log.WithFields(log.Fields{
			"cmd": cmd,
			"args": args,
		}).Fatal("Failed to execute command")
	}
}

func init() {
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

func main() {
	config := viper.GetViper()
	ifname := os.Args[1]

	switch config.GetString("bridges.type") {
	case "linux":
		runCmd("ip", []string{"link", "set", "dev", ifname, "nomaster"})
	case "ovs":
		runCmd("ovs-vsctl", []string{"del-port", config.GetString("bridges.name"), ifname})
	default:
		log.Fatalf("Unknown bridge type")
	}
	runCmd("ip", []string{"link", "set", ifname, "down"})
}
