package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	"github.com/axsh/openvdc/scheduler"
	"github.com/axsh/openvdc/util"
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
	Short: "",
	Long:  ``,
	Run:   execute,
}
var gRPCAddr string
var mesosMasterAddr string
var listenAddr string
var zkAddr string

func init() {
	rootCmd.PersistentFlags().StringVarP(&mesosMasterAddr, "master", "", "localhost:5050", "Mesos Master node address")
	rootCmd.PersistentFlags().StringVarP(&gRPCAddr, "api", "a", "localhost:5000", "gRPC API bind address")
	rootCmd.PersistentFlags().StringVarP(&listenAddr, "listen", "l", "0.0.0.0", "Local bind address")
	rootCmd.PersistentFlags().StringVarP(&zkAddr, "zk", "", "127.0.0.1", "Zookeeper node address")
	rootCmd.PersistentFlags().SetAnnotation("master", cobra.BashCompSubdirsInDir, []string{})
}

func setupDatabaseSchema() {
	ctx, err := model.Connect(context.Background(), []string{zkAddr})
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
	util.SetupLog()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
