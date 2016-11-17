package cmd

import (
        log "github.com/Sirupsen/logrus"
        "github.com/spf13/cobra"
        util "github.com/axsh/openvdc/util"
)

func init() {
        stopCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
        stopCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var stopCmd = &cobra.Command{
        Use:   "stop [Image ID]",
        Short: "Stop an instance",
        Long:  "Stop a running instance.",
        RunE: func(cmd *cobra.Command, args []string) error {

        if len(args) > 0 {
                instanceID := args[0]
                util.SendToApi(serverAddr, instanceID, "", "stop")
        } else { log.Warn("OpenVDC: Please provide an Instance ID.  Usage: stop [Image ID]")
        }
        return nil
}}

