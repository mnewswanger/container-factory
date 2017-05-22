package webserver

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

    "go.mikenewswanger.com/docker-automatic-build/dockerbuild"
)

type WebServer struct {
	DockerBaseDirectory    string
	DockerRegistryBasePath string
	Logger                 *logrus.Logger
	Port                   uint16
	Verbosity              uint8
	ginEngine              *gin.Engine
}

func (ws *WebServer) Serve() {
	ws.Logger.WithFields(logrus.Fields{
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
		Logger:                 ws.Logger,
		Verbosity:              ws.Verbosity,
	}
}
