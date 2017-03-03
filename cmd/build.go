package cmd

import (
	"github.com/spf13/cobra"

	"gitlab.home.mikenewswanger.com/golang/docker-automatic-build/dockerbuild"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build All Docker Images",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var db = dockerbuild.DockerBuild{
			Debug:               commandLineFlags.debug,
			Verbosity:           uint8(commandLineFlags.verbosity),
			DockerfileDirectory: commandLineFlags.dockerfileDirectory,
			InternalImagePrefix: commandLineFlags.dockerRegistryBasePath,
			Tag:                 commandLineFlags.imageTag,
		}
		db.BuildImages(commandLineFlags.forceRebuild)
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolVarP(&commandLineFlags.forceRebuild, "force-rebuild", "f", false, "Force rebuild on all images")
	buildCmd.Flags().StringVarP(&commandLineFlags.imageTag, "image-tag", "t", "", "Tag for docker images")
}