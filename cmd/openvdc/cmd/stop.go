package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [Image ID]",
	Short: "Stop an instance",
	Long:  "Stop a running instance.",
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) > 0 {
			instanceID := args[0]
			util.SendToApi(util.ServerAddr, instanceID, "", "stop")
		} else {
			log.Warn("OpenVDC: Please provide an Instance ID.  Usage: stop [Image ID]")
		}
		return nil
	}}
