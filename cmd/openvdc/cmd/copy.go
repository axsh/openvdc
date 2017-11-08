package cmd

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/cmd/copy"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var copyCmd = &cobra.Command{
	Use:   "copy [File src path] [Instance ID]:[file dest path]",
	Short: "Copy files to and between instances",
	Long:  "Copy files to and between instances",
	Example: `
	% openvdc copy 1.txt i-xxxxxxx:/tmp/1.txt
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			log.Fatalf("Please provide a source path.")
		}
		if len(args) == 1 {
			log.Fatalf("Please provide a destination path.")
		}

		src := args[0]
		dest := args[1]

		p := strings.Split(dest, ":")
		if len(p) < 2 {
			log.Fatalf("Invalid destination path. Please use this format: [Instance ID]:[file dest path]")
		}

		instanceID, instanceDir := p[0], p[1]
		if instanceID == "" {
			log.Fatalf("Invalid Instance ID")
		}
		if instanceDir == "" {
			log.Fatalf("Invalid destination path")
		}

		req := &api.CopyRequest{
			InstanceId: instanceID,
		}

		var res *api.CopyReply

		err := util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)
			var err error
			res, err = c.Copy(context.Background(), req)
			return err
		})

		if err != nil {
			log.WithError(err).Fatal("Disconnected abnormally")
		}

		client, err := copy.NewClient(res)
		if err != nil {
			log.WithError(err).Fatal("Failed to create client")
		}

		err = client.CopyFile(src, instanceDir)

		if err != nil {
			log.WithError(err).Fatal("Failed to copy file")
		}

		return nil
	},
}
