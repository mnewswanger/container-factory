package cmd

import (
	"github.com/spf13/cobra"

	"go.mikenewswanger.com/docker-automatic-build/dockerbuild"
	"go.mikenewswanger.com/docker-automatic-build/webserver"
)

// buildDeploymentCmd represents the build command
var buildDeploymentCmd = &cobra.Command{
	Use:   "build-deployment <deployment-name>",
	Short: "Build a single Docker deployment image",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		webserver.SetLogger(logger)
		webserver.SetVerbosity(uint8(commandLineFlags.verbosity))
		var db = dockerbuild.DockerBuild{
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
