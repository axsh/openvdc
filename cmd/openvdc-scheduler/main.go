package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	"github.com/axsh/openvdc/scheduler"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

// Build time constant variables from -ldflags
var (
	version   string
	sha       string
	builddate string
	goversion string
)

var rootCmd = &cobra.Command{
	Use:   "openvdc-scheduler",
	Short: "Run openvdc-scheduler process",
	Long:  ``,
	Example: `
	"--zk" and "--master" may be one of:
	  "host:port"
		"zk://host1:port1,host2:port2,.../path"

	Auto detect mesos cluster from Zookeeper.
	% openvdc-scheduler --master=zk://localhost/mesos --zk=zk://192.168.1.10

  Explicitly specify the mesos master address.
	% openvdc-scheduler --master=localhost:5050 --zk=localhost:2181
	`,
	Run: execute,
}
var gRPCAddr string
var mesosMasterAddr string
var listenAddr string
var zkAddr backend.ZkEndpoint

func init() {
	zkAddr = backend.ZkEndpoint{Hosts: []string{"localhost"}, Path: "/openvdc"}

	rootCmd.PersistentFlags().StringVarP(&mesosMasterAddr, "master", "", "zk://localhost/mesos", "Mesos Master node address")
	rootCmd.PersistentFlags().StringVarP(&gRPCAddr, "api", "a", "localhost:5000", "gRPC API bind address")
	rootCmd.PersistentFlags().StringVarP(&listenAddr, "listen", "l", "0.0.0.0", "Local bind address")
	rootCmd.PersistentFlags().VarP(&zkAddr, "zk", "", "Zookeeper node address")
	rootCmd.PersistentFlags().SetAnnotation("master", cobra.BashCompSubdirsInDir, []string{})
}

func setupDatabaseSchema() {
	ctx, err := model.Connect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Fatalf("Could not connect to database: %s", zkAddr)
	}
	defer model.Close(ctx)
	ms, ok := model.GetBackendCtx(ctx).(backend.ModelSchema)
	if !ok {
		return
	}

	err = model.InstallSchemas(ms)
	if err != nil {
		log.WithError(err).Fatal("Failed to install schema")
	}
}

func execute(cmd *cobra.Command, args []string) {
	setupDatabaseSchema()
	scheduler.Run(listenAddr, gRPCAddr, mesosMasterAddr, zkAddr)
}

func main() {
	flag.CommandLine.Parse([]string{})
	rootCmd.AddCommand(openvdc.VersionCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
