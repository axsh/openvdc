package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/registry"
	util "github.com/axsh/openvdc/util"
	"github.com/spf13/cobra"
)

var serverAddr string
var imageName string
var hostName string

func init() {
	createCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
	createCmd.PersistentFlags().StringVarP(&hostName, "name", "n", "", "Existing host name")
	createCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

var createCmd = &cobra.Command{
	Use:   "create [Image ID]",
	Short: "Create an instance",
	Long:  `Register and create a new instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			log.Fatalf("Please provide an Image ID.")
		}

		imageSlug := args[0]

		reg, err := setupLocalRegistry()
		if err != nil {
			log.Fatalln(err)
		}
		mi, err := reg.Find(imageSlug)
		if err != nil {
			if err == registry.ErrUnknownImageName {
				log.Fatalf("Image '%s' not found.", imageSlug)
			} else {
				log.Fatalln(err)
			}
		}
		log.Printf("Found image: %s", imageSlug)


		util.SendToApi(serverAddr, mi.Name, hostName, "create")

		return nil
	}}
