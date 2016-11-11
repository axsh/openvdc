package main

import (
	"fmt"
	"os"

	"github.com/axsh/openvdc/scheduler"
	"github.com/axsh/openvdc/util"
	"github.com/spf13/cobra"
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

func init() {
	rootCmd.PersistentFlags().StringVarP(&mesosMasterAddr, "master", "", "localhost:5050", "Mesos Master node address")
	rootCmd.PersistentFlags().StringVarP(&gRPCAddr, "api", "a", "localhost:5000", "gRPC API bind address")
	rootCmd.PersistentFlags().StringVarP(&listenAddr, "listen", "l", "0.0.0.0", "Local bind address")
	rootCmd.PersistentFlags().SetAnnotation("master", cobra.BashCompSubdirsInDir, []string{})
}

func execute(cmd *cobra.Command, args []string) {
	scheduler.Run(listenAddr, gRPCAddr, mesosMasterAddr)
}

func main() {
	util.SetupLog()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
