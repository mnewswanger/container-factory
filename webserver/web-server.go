package webserver

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"go.mikenewswanger.com/docker-automatic-build/dockerbuild"
)

var logger = logrus.New()
var verbosity = uint8(0)

func SetLogger(l *logrus.Logger) {
	logger = l
}

func SetVerbosity(v uint8) {
	verbosity = v
}

type WebServer struct {
	DockerBaseDirectory    string
	DockerRegistryBasePath string
	Port                   uint16
	ginEngine              *gin.Engine
}

func (ws *WebServer) Serve() {
	logger.WithFields(logrus.Fields{
		"port": ws.Port,
	}).Info("Starting web server")
	ws.ginEngine = gin.Default()
	ws.addRoutes()
	ws.ginEngine.Run(":" + strconv.Itoa(int(ws.Port)))
}

func (ws *WebServer) newDockerBuild() dockerbuild.DockerBuild {
	return dockerbuild.DockerBuild{
		DockerBaseDirectory:    ws.DockerBaseDirectory,
		DockerRegistryBasePath: ws.DockerRegistryBasePath,
	}
}
