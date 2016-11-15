package cmd

import (

	log "github.com/Sirupsen/logrus"
	util "github.com/axsh/openvdc/util"
	"github.com/axsh/openvdc/registry"
	"github.com/spf13/cobra"
)

var serverAddr string
var imageName string
var hostName string

func init() {
	runCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:5000", "gRPC API server address")
	runCmd.PersistentFlags().StringVarP(&hostName, "name", "n", "", "Existing host name")
	runCmd.PersistentFlags().SetAnnotation("server", cobra.BashCompSubdirsInDir, []string{})
}

func setupLocalRegistry() (registry.Registry, error) {
	reg := registry.NewGithubRegistry(UserConfDir)
	if !reg.ValidateCache() {
		err := reg.Fetch()
		if err != nil {
			return nil, err
		}
	}

	refresh, err := reg.IsCacheObsolete()
	if err != nil {
		return nil, err
	}
	if refresh {
		log.Infoln("Updating registry cache.")
		err = reg.Fetch()
		if err != nil {
			return nil, err
		}
	}
	return reg, nil
}

var runCmd = &cobra.Command{
	Use:   "run [Image ID]",
	Short: "Run an instance",
	Long:  `Register and start new instance.`,
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


		util.SendToApi(serverAddr, mi.Name, hostName, "run")

		return nil
	}}
