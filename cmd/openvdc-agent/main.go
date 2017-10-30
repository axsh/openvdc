package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api/collector"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	"github.com/axsh/openvdc/resource"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type VDCAgent struct {
	collector resource.ResourceCollector
	nodeInfo  *model.MonitorNode
	conn      *grpc.ClientConn
}

var (
	DefaultConfPath string
	updateInterval  time.Duration = 5
	zkAddr          backend.ZkEndpoint
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
	viper.SetDefault("resource-collector.endpoint", "10.0.100.12:9092")
	viper.SetDefault("resource-collector.mode", "local")
	exitOnErr(initConfig())
}

func main() {
	agent := newVDCAgent()
	zkAddr.Set(viper.GetString("zookeeper.endpoint"))
	exitOnErr(agent.Run())
}

func newVDCAgent() *VDCAgent {
	c, err := resource.NewCollector(viper.GetViper())
	exitOnErr(err)
	node := &model.MonitorNode{
		Id:        "test-monitor2",
		Resources: &model.ComputingResources{},
	}
	return &VDCAgent{
		collector: c,
		nodeInfo:  node,
	}
}

func (a *VDCAgent) Register(ctx context.Context) {
	err := model.Cluster(ctx).Register(a.nodeInfo)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infoln("Registered on OpenVDC cluster service:", a.nodeInfo.Id)
}

func (a *VDCAgent) updateResources() {
	var err error
	copts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
	if a.conn, err = grpc.DialContext(ctx, viper.GetString("resource-collector.endpoint"), copts...); err != nil {
		fmt.Println("failed conn")
	}
	defer a.conn.Close()
	c := collector.NewResourceCollectorClient(a.conn)
	if _, err := c.ReportResources(context.Background(), a.nodeInfo); err != nil {
		fmt.Println("fail api req")
	}
}

func (a *VDCAgent) Run() error {
	ctx, err := model.ClusterConnect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Fatalf("Failed to connect to cluster service %s", zkAddr.String())
	}
	defer func() {
		err := model.ClusterClose(ctx)
		if err != nil {
			log.WithError(err).Error("Failed ClusterClose")
		}
	}()
	ctx, err = model.Connect(ctx, &zkAddr)
	if err != nil {
		log.WithError(err).Fatalf("Failed to connecto to model server: %s", zkAddr.String())
	}
	defer func() {
		if err := model.Close(ctx); err != nil {
			log.WithError(err).Error("Failed model.Close")
		}
	}()
	a.Register(ctx)

	for {
		if err := a.GetResources(); err != nil {
			return err
		}
		a.updateResources()
		time.Sleep(time.Second * updateInterval)
	}
}

func (a *VDCAgent) GetResources() error {
	var err error

	if a.nodeInfo.Resources.Cpu, err = a.collector.GetCpu(); err != nil {
		return err
	}
	if a.nodeInfo.Resources.Memory, err = a.collector.GetMem(); err != nil {
		return err
	}
	if a.nodeInfo.Resources.Storage, err = a.collector.GetDisk(); err != nil {
		return err
	}
	if a.nodeInfo.Resources.Load, err = a.collector.GetLoadAvg(); err != nil {
		return err
	}
	return nil
}
