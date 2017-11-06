package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var copyCmd = &cobra.Command{
	Use:   "copy [File src path] [Instance ID]:/[file dest path]",
	Short: "Copy files to and between instances",
	Long:  "Copy files to and between instances",
	Example: `
	% openvdc copy 1.txt i-xxxxxxx:/tmp/1.txt
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) <= 1 {
			log.Fatalf("Please provide a source path.")
		}
		if len(args) < 2 {
			log.Fatalf("Please provide a destination path.")
		}

		src := args[0]
		dest := args[1]

		req := &api.CopyRequest{
			Src: src,
			Dest: dest,
		}

		return util.RemoteCall(func(conn *grpc.ClientConn) error {
			c := api.NewInstanceClient(conn)
			res, err := c.Copy(context.Background(), req)
			if err != nil {
				log.WithError(err).Fatal("Disconnected abnormaly")
				return err
			}
			fmt.Println(res)
			return err
		})
	},
}
