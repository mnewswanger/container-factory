package cmd

import (
	"github.com/spf13/cobra"
)

// buildDeploymentCmd represents the build command
var buildDeploymentCmd = &cobra.Command{
	Use:   "build-deployment <deployment-name>",
	Short: "Build a single Docker deployment image",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// var db = dockerbuild.DockerBuild{
		// 	Debug:                  commandLineFlags.debug,
		// 	Verbosity:              uint8(commandLineFlags.verbosity),
		// 	DockerfileDirectory:    commandLineFlags.dockerfileDirectory,
		// 	DockerRegistryBasePath: commandLineFlags.dockerRegistryBasePath,
		// 	Tag: commandLineFlags.imageTag,
		// }
		// db.BuildImages(true, !commandLineFlags.localOnly)
	},
}

func init() {
	RootCmd.AddCommand(buildDeploymentCmd)
	buildDeploymentCmd.Flags().BoolVarP(&commandLineFlags.localOnly, "local-only", "l", false, "Skip push build images to upstream repository step")
	buildDeploymentCmd.Flags().StringVarP(&commandLineFlags.imageTag, "image-tag", "t", "", "Tag for docker images")
}
