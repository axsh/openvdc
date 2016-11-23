package cmd

import (
	log "github.com/Sirupsen/logrus"
	util "github.com/axsh/openvdc/util"
	"github.com/spf13/cobra"
)

func init() {
	consoleCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
	consoleCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var consoleCmd = &cobra.Command{
	Use:   "console [Instance ID]",
	Short: "Connect to an instance",
	Long:  "Connect to an instance.",
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) > 0 {
			instanceID := args[0]
			util.SendToApi(serverAddr, instanceID, "", "console")
		} else {
			log.Warn("OpenVDC: Please provide an Instance ID.  Usage: console [Image ID]")
		}
		return nil
	}}
