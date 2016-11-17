package util

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func SetupLog() {
	log.SetOutput(os.Stdout)
}

func SendToApi(serverAddr string, hostName string, imageTitle string, taskType string) {

	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("ERROR: Cannot connect to OpenVDC API: %v", err)
	}

	defer conn.Close()

	c := api.NewInstanceClient(conn)

	resp, err := c.Run(context.Background(), &api.RunRequest{imageTitle, hostName, taskType})

	if err != nil {
		log.Fatalf("ERROR: Cannot connect to OpenVDC API: %v", err)
	}

	log.Println(resp)
}
