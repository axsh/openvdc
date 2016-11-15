package cmd

import (
        log "github.com/Sirupsen/logrus"
        "github.com/spf13/cobra"
        util "github.com/axsh/openvdc/util"
)

func init() {
        startCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
        startCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var startCmd = &cobra.Command{
        Use:   "start [Image ID]",
        Short: "Start an instance",
        Long:  "Start an instance.",
        RunE: func(cmd *cobra.Command, args []string) error {

        if len(args) > 0 {
                instanceID := args[0]
                util.SendToApi(serverAddr, instanceID, "", "start")
        } else { log.Warn("OpenVDC: Please provide an Instance ID.  Usage: start [Image ID]")
        }
        return nil
}}
