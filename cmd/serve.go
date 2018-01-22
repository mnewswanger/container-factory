package cmd

import (
	"github.com/spf13/cobra"

	"go.mikenewswanger.com/container-factory/webserver"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a web service to interact with the build tool",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		webserver.Serve(
			commandLineFlags.dockerBaseDirectory,
			commandLineFlags.dockerRegistryBasePath,
			commandLineFlags.listenPort,
			logger,
			uint8(commandLineFlags.verbosity),
		)
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().Uint16VarP(&commandLineFlags.listenPort, "listen-port", "l", 8080, "Port for web server to listen on")
}
