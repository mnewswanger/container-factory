package cmd

import (
	"sort"

	"github.com/spf13/cobra"

	"go.mikenewswanger.com/docker-automatic-build/dockerbuild"
	"go.mikenewswanger.com/docker-automatic-build/webserver"
)

// listBaseImagesCmd represents the list command
var listDeploymentsCmd = &cobra.Command{
	Use:   "list-deployments",
	Short: "List Configured Deployments",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		webserver.SetLogger(logger)
		webserver.SetVerbosity(uint8(commandLineFlags.verbosity))
		var db = dockerbuild.DockerBuild{
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
