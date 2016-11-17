package cmd

import (
        log "github.com/Sirupsen/logrus"
        "github.com/spf13/cobra"
        util "github.com/axsh/openvdc/util"
)

func init() {
        runCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
        runCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var runCmd = &cobra.Command{
        Use:   "run [Image ID]",
        Short: "Run an instance",
        Long:  "Run an instance.",
        RunE: func(cmd *cobra.Command, args []string) error {

        if len(args) > 0 {
                instanceID := args[0]
                util.SendToApi(serverAddr, instanceID, "", "run")
        } else { log.Warn("OpenVDC: Please provide an Instance ID.  Usage: run [Image ID]")
        }
        return nil
}}
