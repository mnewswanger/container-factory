package cmd

import (
	"github.com/spf13/cobra"

	"gitlab.home.mikenewswanger.com/golang/docker-automatic-build/dockerbuild"
)

// listBaseImagesCmd represents the list command
var listDeploymentsCmd = &cobra.Command{
	Use:   "list-deployments",
	Short: "List Configured Deployments",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var db = dockerbuild.DockerBuild{
			Debug:                  commandLineFlags.debug,
			Verbosity:              uint8(commandLineFlags.verbosity),
			DockerBaseDirectory:    commandLineFlags.dockerBaseDirectory,
			DockerRegistryBasePath: commandLineFlags.dockerRegistryBasePath,
		}
		db.PrintDeployments()
	},
}

func init() {
	RootCmd.AddCommand(listDeploymentsCmd)
}
