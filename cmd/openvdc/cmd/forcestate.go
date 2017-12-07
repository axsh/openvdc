package cmd

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/axsh/openvdc/model"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var forceStateCmd = &cobra.Command{
	Use:   "forcestate [Instance ID] [State]",
	Short: "Forcefully sets instance to specified state",
	Long:  "Forcefully sets instance to specified state",
	Example: `
	% openvdc forcestate i-0000000001 running
	`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			log.Fatalf("Please provide an instance ID.")
		}

		if len(args) == 1 {
			log.Fatalf("Please provide a desired instance state.")
		}

		instanceID := args[0]
		goalState, ok := model.InstanceState_State_value[strings.ToUpper(args[1])]

		return nil
	},
}
