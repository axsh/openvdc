package cmd

import (
	"context"
	"log"

	pb "github.com/axsh/openvdc/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var serverAddr string

func init() {
	runCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
	runCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var runCmd = &cobra.Command{
	Use:   "run [Image ID]",
	Short: "Run an instance",
	Long:  `Register and start new instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		// Todo: fetch remote json
                viper.SetConfigType("json")
                viper.SetConfigName("1box-centos7")
                viper.AddConfigPath("../../deployment/1box")
                err := viper.ReadInConfig()

                if err != nil {
                        log.Fatalf("Error reading config file: ", err)
                }

                result := viper.GetString("variables.memory")

		conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Connection error: %v", err)
		}

		defer conn.Close()

		c := pb.NewInstanceClient(conn)

		resp, err := c.Run(context.Background(), &pb.RunRequest{result})
		if err != nil {
			log.Fatalf("RPC error: %v", err)
		}
		log.Println(resp)
		return nil
	},
}
