package main

import "github.com/axsh/openvdc/cmd/openvdc/cmd"

// Build time constant variables from -ldflags
var (
	version   string
	sha       string
	builddate string
	goversion string
)

func main() {
	cmd.Execute()
}
