package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	"github.com/axsh/openvdc/scheduler"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

var DefaultConfPath string

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

func init() {
	viper.SetDefault("mesos.master", "zk://localhost/mesos")
	viper.SetDefault("mesos.listen", "0.0.0.0")
	viper.SetDefault("zookeeper.endpoint", "zk://localhost/openvdc")
	viper.SetDefault("api.endpoint", "localhost:5000")

	cobra.OnInitialize(initConfig)
	pfs := rootCmd.PersistentFlags()
	pfs.String("config", DefaultConfPath, "Load config file from the path")

	pfs.String("master", viper.GetString("mesos.master"), "Mesos Master node address")
	viper.BindPFlag("mesos.master", pfs.Lookup("master"))
	pfs.String("listen", viper.GetString("mesos.listen"), "Mesos scheduler local bind address")
	viper.BindPFlag("mesos.listen", pfs.Lookup("listen"))
	pfs.String("api", viper.GetString("api.endpoint"), "gRPC API bind address")
	viper.BindPFlag("api.endpoint", pfs.Lookup("api"))
	pfs.String("zk", viper.GetString("zookeeper.endpoint"), "Zookeeper node address")
	viper.BindPFlag("zookeeper.endpoint", pfs.Lookup("zk"))
}

func setupDatabaseSchema() {
	var zkAddr backend.ZkEndpoint
	zkAddr.Set(viper.GetString("zookeeper.endpoint"))
	ctx, err := model.Connect(context.Background(), zkAddr)
	if err != nil {
		log.WithError(err).Fatalf("Could not connect to database: %s", zkAddr.String())
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

func initConfig() {
	f := rootCmd.PersistentFlags().Lookup("config")
	if f.Changed {
		viper.SetConfigFile(f.Value.String())
	} else {
		viper.SetConfigFile(DefaultConfPath)
		viper.SetConfigType("toml")
	}
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		if viper.ConfigFileUsed() == DefaultConfPath && os.IsNotExist(err) {
			// Ignore default conf file does not exist.
			return
		}
		log.Fatalf("Failed to load config %s: %v", viper.ConfigFileUsed(), err)
	}
}

func execute(cmd *cobra.Command, args []string) {
	setupDatabaseSchema()
	var zkAddr backend.ZkEndpoint
	zkAddr.Set(viper.GetString("zookeeper.endpoint"))
	scheduler.Run(
		viper.GetString("mesos.listen"),
		viper.GetString("api.endpoint"),
		viper.GetString("mesos.master"),
		zkAddr)
}

func main() {
	flag.CommandLine.Parse([]string{})
	log.SetFormatter(&cmd.LogFormatter{})
	rootCmd.AddCommand(cmd.VersionCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
