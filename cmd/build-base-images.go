package cmd

import (
	"github.com/spf13/cobra"

	"go.mikenewswanger.com/docker-automatic-build/dockerbuild"
	"go.mikenewswanger.com/docker-automatic-build/webserver"
)

// buildBaseImagesCmd represents the build command
var buildBaseImagesCmd = &cobra.Command{
	Use:   "build-base-images",
	Short: "Build All Docker Images",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		webserver.SetLogger(logger)
		webserver.SetVerbosity(uint8(commandLineFlags.verbosity))
		var db = dockerbuild.DockerBuild{
			DockerBaseDirectory:    commandLineFlags.dockerBaseDirectory,
			DockerRegistryBasePath: commandLineFlags.dockerRegistryBasePath,
			Tag: commandLineFlags.imageTag,
		}
		db.BuildBaseImages(commandLineFlags.forceRebuild, !commandLineFlags.localOnly)
	},
}

func init() {
	RootCmd.AddCommand(buildBaseImagesCmd)
	buildBaseImagesCmd.Flags().BoolVarP(&commandLineFlags.forceRebuild, "force-rebuild", "f", false, "Force rebuild on all images")
	buildBaseImagesCmd.Flags().BoolVarP(&commandLineFlags.localOnly, "local-only", "l", false, "Skip push build images to upstream repository step")
	buildBaseImagesCmd.Flags().StringVarP(&commandLineFlags.imageTag, "image-tag", "t", "", "Tag for docker images")
}
