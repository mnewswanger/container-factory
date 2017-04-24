package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "docker-automatic-build",
	Short: "Docker Automated Build Tool",
	Long:  ``,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&commandLineFlags.dockerRegistryBasePath, "registry-base-path", "p", "", "Docker Registry Base Path i.e. registry.example.com/")
	RootCmd.PersistentFlags().StringVarP(&commandLineFlags.dockerfileDirectory, "dockerfile-directory", "d", "", "Docker Build Directory")
	RootCmd.PersistentFlags().CountVarP(&commandLineFlags.verbosity, "verbosity", "v", "Output verbosity")
	RootCmd.PersistentFlags().BoolVarP(&commandLineFlags.debug, "debug", "", false, "Debug level output")
}
