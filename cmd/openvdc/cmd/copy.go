package cmd

import (
	"fmt"
	"net"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/cmd/openvdc/internal/util"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/crypto/ssh"
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
		if len(args) == 0 {
			log.Fatalf("Please provide a source path.")
		}
		if len(args) == 1 {
			log.Fatalf("Please provide a destination path.")
		}

		//src := args[0]
		dest := args[1]

		path := strings.Split(dest, ":")
        	fmt.Sprintf("value: %s", path[0])

        	instanceID, instanceDir := path[0], path[1]

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

		host, port, err := net.SplitHostPort(res.GetAddress())
               	if err != nil {
                log.Fatal("Invalid ssh host address: ", res.GetAddress())
                }

		clientConfig := &ssh.ClientConfig{
                	User: instanceID,
        	}
		
		client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), clientConfig)

		if err != nil {
			log.WithError(err).Fatal("ssh.Dial failed")
		}
		
		session, err := client.NewSession()
		if err != nil {
			log.WithError(err).Fatal("Failed to create session")
		}
		defer session.Close()

		go func() {
                	//Todo: copy file
        	}()


		err = session.Run("/usr/bin/scp -tr ./")
		if err != nil {
			log.WithError(err).Fatal("session.Run failed")
		}

		return nil
	},
}
