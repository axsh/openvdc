package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy [Image ID]",
	Short: "Destroy an instance",
	Long:  "Destroy an already existing instance.",
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) > 0 {
			instanceID := args[0]
			util.SendToApi(util.ServerAddr, instanceID, "", "destroy")
		} else {
			log.Warn("OpenVDC: Please provide an Instance ID.  Usage: destroy [Image ID]")
		}
		return nil
	}}
