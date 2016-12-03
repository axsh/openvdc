package cmd

import (
	"fmt"
	"os"

	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
)

var cfgFile string

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
	RootCmd.AddCommand(registerCmd)
	RootCmd.AddCommand(unregisterCmd)
	RootCmd.AddCommand(startCmd)
	RootCmd.AddCommand(showCmd)
	RootCmd.AddCommand(TemplateCmd)
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.openvdc.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.PersistentFlags().StringVarP(&util.ServerAddr, "server", "s", "localhost:5000", "gRPC API server address")
	RootCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}
