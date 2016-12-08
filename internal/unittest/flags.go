package unittest

import "os"

// Place global variables or flags only for unit test code.

var TestZkServer = "127.0.0.1:2181"
var GithubDefaultRef = "master"

func init() {
	if v, exists := os.LookupEnv("ZK"); exists {
		TestZkServer = v
	}
	if v, exists := os.LookupEnv("GITHUB_DEFAULT_REF"); exists {
		GithubDefaultRef = v
	}
}
