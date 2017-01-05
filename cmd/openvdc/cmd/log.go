package cmd

import (
	"fmt"
	"strconv"
	"strings"

	mlog "github.com/ContainX/go-mesoslog/mesoslog"
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
)

var tail bool

func init() {
	logCmd.Flags().BoolVarP(&tail, "tail", "t", false, "Tail log output instead of just printing it once.")
}

var logCmd = &cobra.Command{
	Use:   "log [Instance ID]",
	Short: "Print logs of an instance",
	Long:  "Print logs of an instance",
	Example: `
	% openvdc log i-xxxxxxx
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			log.Fatalf("Please provide an Instance ID.")
		}

		instanceID := "VDC_" + args[0]

		split := strings.Split(util.MesosMasterAddr, ":")
		mesosMasterAddr := split[0]
		mesosMasterPort, err := strconv.Atoi(split[1])

		if err != nil {
			log.WithError(err).Fatal("Error trying to convert string")
			return err
		}

		if tail == false {
			cl, err := mlog.NewMesosClientWithOptions(mesosMasterAddr, mesosMasterPort, &mlog.MesosClientOptions{SearchCompletedTasks: false, ShowLatestOnly: true})
			if err != nil {
				log.WithError(err).Fatal("Couldn't connect to Mesos master")
				return err
			}

			result, err := cl.GetLog(instanceID, mlog.STDERR, "")
			if err != nil {
				log.WithError(err).Fatal("Error fetching log")
				return err
			}

			for _, log := range result {
				fmt.Printf(log.Log)
			}
		} else {
			cl, err := mlog.NewMesosClientWithOptions(mesosMasterAddr, mesosMasterPort, nil)

			if err != nil {
				log.WithError(err).Fatal("Couldn't connect to Mesos master")
				return err
			}

			err = cl.TailLog(instanceID, mlog.STDOUT, 5)

			if err != nil {
				log.WithError(err).Fatal("Error trying to tail log")
				return err
			}
		}
		return err
	},
}
