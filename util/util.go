package util

import (
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
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
