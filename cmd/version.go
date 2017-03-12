package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Build time constant variables from -ldflags
var (
	Version   string
	Sha       string
	Builddate string
	Goversion string
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Long:  "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s (git: %s, build: %s, goversion: %s)\n",
			cmd.Root().CommandPath(),
			Version,
			Sha[:8],
			Builddate,
			Goversion,
		)
	},
}
