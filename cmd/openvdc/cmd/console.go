package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	// TODO: Remove --server option from sub-command.
	consoleCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
	consoleCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var consoleCmd = &cobra.Command{
	Use:   "console [Instance ID]",
	Short: "Connect to an instance",
	Long:  "Connect to an instance.",
	RunE: func(cmd *cobra.Command, args []string) error {

				
		log.Infoln("Under construction")


		return nil
	},
}
