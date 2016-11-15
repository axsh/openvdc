package util

import (
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func GetRemoteJsonField(field string, url string) string {
	r, err := http.Get(url)

	if err != nil {
		log.Fatal("Couldn't fetch remote JSON: ", err)
	} else {
		defer r.Body.Close()
	}

	viper.SetConfigType("json")
	viper.ReadConfig(r.Body)

	result := viper.GetString(field)

	return result
}

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
