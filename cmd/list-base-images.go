package cmd

import (
	"github.com/spf13/cobra"

	"gitlab.home.mikenewswanger.com/infrastructure/docker-automatic-build/dockerbuild"
)

// listBaseImagesCmd represents the list command
var listBaseImagesCmd = &cobra.Command{
	Use:   "list-base-images",
	Short: "List Dockerfile Heirarchy",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var db = dockerbuild.DockerBuild{
			Debug:                  commandLineFlags.debug,
			Verbosity:              uint8(commandLineFlags.verbosity),
			DockerBaseDirectory:    commandLineFlags.dockerBaseDirectory,
			DockerRegistryBasePath: commandLineFlags.dockerRegistryBasePath,
		}
		db.PrintBaseImageHeirarchy()
	},
}

func init() {
	RootCmd.AddCommand(listBaseImagesCmd)
}
