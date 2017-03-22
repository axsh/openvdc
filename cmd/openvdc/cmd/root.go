package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "openvdc",
	Short: "",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	RootCmd.AddCommand(createCmd)
	RootCmd.AddCommand(destroyCmd)
	RootCmd.AddCommand(runCmd)
	RootCmd.AddCommand(stopCmd)
	RootCmd.AddCommand(consoleCmd)
	RootCmd.AddCommand(startCmd)
	RootCmd.AddCommand(showCmd)
	RootCmd.AddCommand(TemplateCmd)
	RootCmd.AddCommand(logCmd)
	RootCmd.AddCommand(rebootCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(cmd.VersionCmd)
	RootCmd.AddCommand(waitCmd)
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.
	viper.SetDefault("api.endpoint", "127.0.0.1:5000")

	pfs := RootCmd.PersistentFlags()
	pfs.String("config", filepath.Join(util.UserConfDir, "config.toml"), "Load config file from the path")
	pfs.String("server", viper.GetString("api.endpoint"), "gRPC API server address")
	pfs.SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
	viper.BindPFlag("api.endpoint", pfs.Lookup("server"))
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	f := RootCmd.PersistentFlags().Lookup("config")
	if f.Changed {
		viper.SetConfigFile(f.Value.String())
	}
	viper.SetConfigName("config")
	viper.AddConfigPath(filepath.Join(util.UserConfDir))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to load config %s: %v", viper.ConfigFileUsed(), err)
	}
}
