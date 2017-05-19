package cmd

import (
	"sort"

	"github.com/spf13/cobra"

	"gitlab.home.mikenewswanger.com/infrastructure/docker-automatic-build/dockerbuild"
)

// listBaseImagesCmd represents the list command
var listDeploymentsCmd = &cobra.Command{
	Use:   "list-deployments",
	Short: "List Configured Deployments",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var db = dockerbuild.DockerBuild{
			Verbosity:              uint8(commandLineFlags.verbosity),
			DockerBaseDirectory:    commandLineFlags.dockerBaseDirectory,
			DockerRegistryBasePath: commandLineFlags.dockerRegistryBasePath,
		}
		var deployments = db.GetDeployments()
		sort.Strings(deployments)
		for _, d := range deployments {
			println(d)
		}
	},
}

func init() {
	RootCmd.AddCommand(listDeploymentsCmd)
}
