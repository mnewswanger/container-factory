package webserver

import (
	"github.com/gin-gonic/gin"
	"gitlab.home.mikenewswanger.com/infrastructure/docker-automatic-build/dockerbuild"
)

func (ws *WebServer) addRoutes() {
	ws.ginEngine.GET("/base-images/build", func(c *gin.Context) { ws.buildBaseImages(c) })
	ws.ginEngine.GET("/base-images/list", func(c *gin.Context) { ws.renderBaseImagesList(c) })
	ws.ginEngine.GET("/deployments/build", func(c *gin.Context) { ws.buildDeployment(c) })
	ws.ginEngine.GET("/deployments/list", func(c *gin.Context) { ws.renderDeploymentsList(c) })
}

func (ws *WebServer) buildBaseImages(c *gin.Context) {
	var db = ws.newDockerBuild()
	db.Tag = c.Query("tag")
	c.String(200, "Build process started")
	go func(db *dockerbuild.DockerBuild, forceRebuild bool) {
		db.BuildBaseImages(forceRebuild, true)
	}(&db, c.Query("force-rebuild") != "")
}

func (ws *WebServer) buildDeployment(c *gin.Context) {
	var db = ws.newDockerBuild()
	db.Tag = c.Query("tag")
	db.DeploymentTag = c.Query("deployment-tag")
	c.String(200, "Build process started")
	go func(db *dockerbuild.DockerBuild, deploymentName string) {
		db.BuildDeployment(deploymentName, true)
	}(&db, c.Query("name"))
}

func (ws *WebServer) renderBaseImagesList(c *gin.Context) {
	var db = ws.newDockerBuild()
	var buildableImages, orphanedImages = db.GetBaseImageHeirarchy()

	switch c.Query("format") {
	case "json":
		c.JSON(200, gin.H{
			"buildable_images": buildableImages,
			"orphaned_images":  orphanedImages,
		})
	case "yaml":
		c.YAML(200, gin.H{
			"buildable_images": buildableImages,
			"orphaned_images":  orphanedImages,
		})
	default:
		var output = "Buildable Images:\n"
		for _, bi := range buildableImages {
			output += "  " + bi.Name + "\n" + getImageChildrenString(bi, "  ")
		}
		output += "\n"
		output += "Orphaned Images\n"
		for _, oi := range orphanedImages {
			output += oi.Name + " (requested from " + oi.ParentName + ")\n"
		}
		c.String(200, output)
	}
}

func (ws *WebServer) renderDeploymentsList(c *gin.Context) {
	var db = ws.newDockerBuild()
	var deployments = db.GetDeployments()

	switch c.Query("format") {
	case "json":
		c.JSON(200, gin.H{
			"deployments": deployments,
		})
	case "yaml":
		c.YAML(200, gin.H{
			"deployments": deployments,
		})
	default:
		var output string
		for _, d := range deployments {
			output += d + "\n"
		}
		c.String(200, output)
	}
}

func getImageChildrenString(dbi dockerbuild.DockerBuildableImage, prefix string) string {
	var imageChildrenString string
	for _, c := range dbi.Children {
		imageChildrenString += prefix + "  ↳" + c.Name + "\n"
		imageChildrenString += getImageChildrenString(c, "   "+prefix)
	}
	return imageChildrenString
}
