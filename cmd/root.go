package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type flags struct {
	verbosity              int
	deploymentImageTag     string
	dockerBaseDirectory    string
	dockerRegistryBasePath string
	forceRebuild           bool
	imageTag               string
	listenPort             uint16
	localOnly              bool
	outputFormat           string
}

var commandLineFlags = flags{}

var logger = logrus.New()

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "container-factory",
	Short: "Container Image Build Tool",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		switch commandLineFlags.verbosity {
		case 0:
			logger.Level = logrus.ErrorLevel
			break
		case 1:
			logger.Level = logrus.WarnLevel
			break
		case 2:
			fallthrough
		case 3:
			logger.Level = logrus.InfoLevel
			break
		default:
			logger.Level = logrus.DebugLevel
			break
		}

		logger.Debug("Pre-run complete")
	},
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
	RootCmd.PersistentFlags().StringVarP(&commandLineFlags.dockerRegistryBasePath, "registry-base-path", "p", "", "Image Registry Base Path i.e. registry.example.com")
	RootCmd.PersistentFlags().StringVarP(&commandLineFlags.dockerBaseDirectory, "digest-base-directory", "d", "", "Base Directory for build assets")
	RootCmd.PersistentFlags().CountVarP(&commandLineFlags.verbosity, "verbosity", "v", "Output verbosity")
}
