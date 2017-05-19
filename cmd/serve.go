package cmd

import (
	"github.com/spf13/cobra"
	webserver "gitlab.home.mikenewswanger.com/infrastructure/docker-automatic-build/webserver"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a web service to interact with the build tool",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var ws = webserver.WebServer{
			Logger:                 logger,
			DockerBaseDirectory:    commandLineFlags.dockerBaseDirectory,
			DockerRegistryBasePath: commandLineFlags.dockerRegistryBasePath,
			Port:      commandLineFlags.listenPort,
			Verbosity: uint8(commandLineFlags.verbosity),
		}
		ws.Serve()
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().Uint16VarP(&commandLineFlags.listenPort, "listen-port", "l", 8080, "Port for web server to listen on")
}
