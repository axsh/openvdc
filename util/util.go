package util

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

func SetupLog() {
	log.SetOutput(os.Stdout)
}
