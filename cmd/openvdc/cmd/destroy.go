package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy [Instance ID]",
	Short: "Destroy an instance",
	Long:  "Destroy an already existing instance.",
	RunE: func(cmd *cobra.Command, args []string) error {

		log.Infoln("Under construction")

		return nil
	},
}
