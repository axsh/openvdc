package main

import (
	"fmt"
	"os"
	"time"
	"net"

	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/resources"
	"github.com/axsh/openvdc/api/agent"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"
)

type VDCAgent struct {
	collector resources.ResourceCollector
	resources *model.ComputingResources
	AgentId   string
	listener  net.Listener
}

var (
	vdcAgent *VDCAgent
	DefaultConfPath string
	updateInteval time.Duration = 5
	// TODO: receive from configuration file or as flag
)

func initConfig() error {
	viper.SetConfigFile(DefaultConfPath)
	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		if viper.ConfigFileUsed() == DefaultConfPath && os.IsNotExist(err) {
			// Ignore default conf file does not exist.
			return nil
		}
		return err
	}
	return nil
}

func exitOnErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	viper.SetDefault("resource-collector.listen", "0.0.0.0:9092")
	viper.SetDefault("resource-collector.advertise-ip", "")
	viper.SetDefault("resource-collector.mode", "local")

	exitOnErr(initConfig())

}

func main() {
	vdcAgent = newVDCAgent()
	exitOnErr(vdcAgent.Run())
}

func newVDCAgent() *VDCAgent {
	c, err := resources.NewCollector(viper.GetViper());
	exitOnErr(err)

	l, err := net.Listen("tcp", viper.GetString("resource-collector.listen"))
	exitOnErr(err)

	return &VDCAgent{
		collector: c,
		resources: &model.ComputingResources{},
		listener: l,
	}
}

func (a *VDCAgent) startAgentAPIServer() *agent.AgentAPIServer {
	s := agent.NewAgentAPIServer()
	go s.Serve(a.listener)
	return s
}

func (a *VDCAgent) Run() error {
	s := vdcAgent.startAgentAPIServer()
	defer s.GracefulStop()

	for {
		if err := a.GetResources(); err != nil {
			return err
		}

		fmt.Println(a.resources)
		message, err := proto.Marshal(a.resources)
		if err != nil {
			return err
		}

		fmt.Println(message)
		time.Sleep(time.Second * updateInteval)
	}
}

func (a *VDCAgent) GetResources() error {
	var err error

	if a.resources.Cpu, err = a.collector.GetCpu(); err != nil {
		return err
	}
	if a.resources.Memory, err = a.collector.GetMem(); err != nil {
		return err
	}
	if a.resources.Storage, err = a.collector.GetDisk(); err != nil {
		return err
	}
	if a.resources.Load, err = a.collector.GetLoadAvg(); err != nil {
		return err
	}
	return nil
}
