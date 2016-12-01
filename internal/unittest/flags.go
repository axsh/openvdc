package unittest

import (
	"os"
)

// Place global variables or flags only for unit test code.

var TestZkServer = "127.0.0.1:2181"

func init() {
	if v, exists := os.LookupEnv("ZK"); exists {
		TestZkServer = v
	}
}
