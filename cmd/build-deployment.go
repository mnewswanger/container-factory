package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.home.mikenewswanger.com/infrastructure/docker-automatic-build/dockerbuild"
)

// buildDeploymentCmd represents the build command
var buildDeploymentCmd = &cobra.Command{
	Use:   "build-deployment <deployment-name>",
	Short: "Build a single Docker deployment image",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var db = dockerbuild.DockerBuild{
			Verbosity:              uint8(commandLineFlags.verbosity),
			DockerBaseDirectory:    commandLineFlags.dockerBaseDirectory,
			DockerRegistryBasePath: commandLineFlags.dockerRegistryBasePath,
			Tag:           commandLineFlags.imageTag,
			DeploymentTag: commandLineFlags.deploymentImageTag,
		}
		db.BuildDeployment(args[0], !commandLineFlags.localOnly)
	},
}

func init() {
	RootCmd.AddCommand(buildDeploymentCmd)
	buildDeploymentCmd.Flags().BoolVarP(&commandLineFlags.localOnly, "local-only", "l", false, "Skip push build images to upstream repository step")
	buildDeploymentCmd.Flags().StringVarP(&commandLineFlags.deploymentImageTag, "deployment-image-tag", "", "", "Tag for docker deployment")
	buildDeploymentCmd.Flags().StringVarP(&commandLineFlags.imageTag, "base-image-tag", "t", "", "Tag for docker images during deployment build process")
}
