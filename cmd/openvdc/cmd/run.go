package cmd

import (
	"context"
	"log"
	"fmt"

	pb "github.com/axsh/openvdc/proto"
	util "github.com/axsh/openvdc/util"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var serverAddr string
var imageName string
var hostName string

func init() {
	runCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
	runCmd.PersistentFlags().StringVarP(&imageName,"image", "i", "centos7", "Name of image file")
        runCmd.PersistentFlags().StringVarP(&hostName,"name", "n", "", "Existing host name")
	runCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var runCmd = &cobra.Command{
	Use:   "run [Image ID]",
	Short: "Run an instance",
	Long:  `Register and start new instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		result := util.GetRemoteJsonField("variables.memory", "https://raw.githubusercontent.com/axsh/openvdc/master/deployment/1box/1box-centos7.json")

		fmt.Println("Json: ", result)
                fmt.Println("Image name: ", imageName)
                fmt.Println("Host name: ", hostName)

		conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Connection error: %v", err)
		}

		defer conn.Close()

		c := pb.NewInstanceClient(conn)

		resp, err := c.Run(context.Background(), &pb.RunRequest{imageName,hostName})
		if err != nil {
			log.Fatalf("RPC error: %v", err)
		}
		log.Println(resp)
		return nil
	},
}
