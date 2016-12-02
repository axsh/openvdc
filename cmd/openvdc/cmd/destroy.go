package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	// TODO: Remove --server option from sub-command.
	destroyCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
	destroyCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var destroyCmd = &cobra.Command{
	Use:   "destroy [Instance ID]",
	Short: "Destroy an instance",
	Long:  "Destroy an already existing instance.",
	RunE: func(cmd *cobra.Command, args []string) error {

		log.Infoln("Under construction")

		return nil
	},
}
