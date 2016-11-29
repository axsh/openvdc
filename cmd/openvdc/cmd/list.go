package cmd

import (
	log "github.com/Sirupsen/logrus"
	//util "github.com/axsh/openvdc/util"
	"github.com/spf13/cobra"
	"github.com/axsh/openvdc/model"
	"golang.org/x/net/context"
)

func init() {
	// TODO: Remove --server option from sub-command.
	listCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
	listCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

func getInstances(filter string, ctx context.Context) (string) {

	var instanceFilter model.Instance_State

	switch(filter) {
		case "registered":
			instanceFilter = model.Instance_REGISTERED

		case "running":
			instanceFilter = model.Instance_RUNNING
	}

	instances, err := model.Instances(ctx).FilterByState(instanceFilter)

	if err != nil {
		log.Errorln("Error: ", err)
	}

	var result string

	if len(instances) > 0 {
		for _, i := range instances {
			result = result + " " + i.GetId()
		}
	} 
		
	return result
}


var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered/running instances",
	Long:  "List all registered/running instances",
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx, err := model.Connect(context.Background(), []string{"127.0.0.1"})
		if err != nil {
			log.WithError(err).Error("Failed to connect to datasource:")
		} else {
			defer model.Close(ctx)
		}

		registeredInstances := getInstances("registered", ctx)
		log.Infoln("Registered instances:\n" + registeredInstances)

		runningInstances := getInstances("running", ctx)
                log.Infoln("Running instances:\n" + runningInstances)

		return nil
	}}
