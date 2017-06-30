package webserver

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"go.mikenewswanger.com/docker-automatic-build/dockerbuild"
)

var ginEngine *gin.Engine
var logger = logrus.New()
var verbosity = uint8(0)
var registryBasePath string

// Serve starts up a webserver
func Serve(dockerBaseDirectory string, dockerRegistryBasePath string, listenPort uint16, l *logrus.Logger, v uint8) {
	logger = l
	verbosity = v
	ginEngine = gin.Default()
	registryBasePath = dockerRegistryBasePath

	dockerbuild.SetLogger(logger)
	dockerbuild.SetVerbosity(verbosity)
	dockerbuild.SetDockerBaseDirectory(dockerBaseDirectory)
	logger.WithFields(logrus.Fields{
		"port": listenPort,
	}).Info("Starting web server")
	addRoutes()
	ginEngine.Run(":" + strconv.Itoa(int(listenPort)))
}
