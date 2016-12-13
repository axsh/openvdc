package cmd

import (
	"github.com/axsh/openvdc/cmd/openvdc/cmd/template"
	"github.com/spf13/cobra"
)

func init() {
	TemplateCmd.AddCommand(template.ValidateCmd)
	TemplateCmd.AddCommand(template.ShowCmd)
}

var TemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "Operations for resource template",
	Long:  "Operations for resource template",
}
