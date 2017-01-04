package cmd

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
)

var tail bool

var rebootCmd = &cobra.Command{
	Use:   "reboot [Instance ID]",
	Short: "Reboots an instance",
	Long:  "Reboots an instance",
	Example: `
	% openvdc reboot i-xxxxxxx
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			log.Fatalf("Please provide an Instance ID.")
		}

		instanceID := args[0]

	},
}
