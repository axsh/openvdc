package cmd

import (
	"context"
	"log"
	"strings"

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
        runCmd.PersistentFlags().StringVarP(&hostName,"name", "n", "", "Existing host name")
        runCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var runCmd = &cobra.Command{
        Use:   "run [Image ID]",
        Short: "Run an instance",
        Long:  `Register and start new instance.`,
        RunE: func(cmd *cobra.Command, args []string) error {

                imageTitle := strings.ToLower(util.GetRemoteJsonField("title", "https://raw.githubusercontent.com/openvdc/images/master/centos-7.json"))
                imageUrl := util.GetRemoteJsonField("images.download_url", "https://raw.githubusercontent.com/openvdc/images/master/centos-7.json")

                if(len(args) > 0 ){
                        inputImageTitle := args[0]

                        if inputImageTitle  ==  imageTitle {

                                log.Println("Found image: ", imageTitle)
                                log.Println("ImageUrl: ", imageUrl)

                                //Viper isn't able to fetch the field "images.download_url" right now. Probably due to square brackets in the Json-file.
                                imageUrl = "https://images.linuxcontainers.org/1.0/images/d767cfe9a0df0b2213e28b39b61e8f79cb9b1e745eeed98c22bc5236f277309a/export"

                                log.Println(imageUrl)

                                conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
                                if err != nil {
                                        log.Fatalf("ERROR: Cannot connect to OpenVDC API: %v", err)
                                }

                                defer conn.Close()

                                c := pb.NewInstanceClient(conn)

                                resp, err := c.Run(context.Background(), &pb.RunRequest{imageTitle,imageUrl,hostName})

                                if err != nil {
                                        log.Fatalf("ERROR: Cannot connect to OpenVDC API: %v", err)
                                }

                                log.Println(resp)
			} else { log.Warn("OpenVDC: Image not found. Available images: centos7")  }
                } else {
                        log.Warn("OpenVDC: Please provide an Image ID.  Usage: run [Image ID]")
        }
                return nil
}}
