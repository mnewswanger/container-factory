package cmd

import (
	"sort"

	"github.com/spf13/cobra"

	"go.mikenewswanger.com/container-factory/dockerbuild"
)

// listBaseImagesCmd represents the list command
var listDeploymentsCmd = &cobra.Command{
	Use:   "list-deployments",
	Short: "List Configured Deployments",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		dockerbuild.SetLogger(logger)
		dockerbuild.SetVerbosity(uint8(commandLineFlags.verbosity))
		dockerbuild.SetDockerBaseDirectory(commandLineFlags.dockerBaseDirectory)
		deployments := dockerbuild.GetDeployments()
		sort.Strings(deployments)
		for _, d := range deployments {
			println(d)
		}
	},
}

func init() {
	RootCmd.AddCommand(listDeploymentsCmd)
}
