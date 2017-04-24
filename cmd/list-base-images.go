package cmd

import (
	"github.com/spf13/cobra"

	"gitlab.home.mikenewswanger.com/golang/docker-automatic-build/dockerbuild"
	"gitlab.home.mikenewswanger.com/golang/filesystem"
)

// listBaseImagesCmd represents the list command
var listBaseImagesCmd = &cobra.Command{
	Use:   "list",
	Short: "List Dockerfile Heirarchy",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var db = dockerbuild.DockerBuild{
			Debug:                  commandLineFlags.debug,
			Verbosity:              uint8(commandLineFlags.verbosity),
			DockerfileDirectory:    filesystem.ForceTrailingSlash(commandLineFlags.dockerfileDirectory) + "dockerfiles",
			DockerRegistryBasePath: commandLineFlags.dockerRegistryBasePath,
		}
		db.PrintImageHeirarchy()
	},
}

func init() {
	RootCmd.AddCommand(listBaseImagesCmd)
}
