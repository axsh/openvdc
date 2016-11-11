package util

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
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

func SetupLog(logpath string, filename string) {

	if _, err := os.Stat(logpath); os.IsNotExist(err) {
		os.Mkdir(logpath, os.ModePerm)
	}

	if _, err := os.Stat(logpath + filename); os.IsNotExist(err) {
		_, err := os.Create(logpath + filename)
		if err != nil {
			log.Println("Error creating log file: ", err)
		}
	}

	vdclog, err := os.OpenFile(logpath+filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("ERROR: Couldn't open log file", err)
	}

	log.SetOutput(vdclog)
}
