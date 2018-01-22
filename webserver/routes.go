package webserver

import (
	"github.com/gin-gonic/gin"

	"go.mikenewswanger.com/container-factory/dockerbuild"
)

func addRoutes() {
	ginEngine.GET("/api/v1/base-images/build", func(c *gin.Context) { buildBaseImages(c) })
	ginEngine.GET("/api/v1/base-images/list", func(c *gin.Context) { renderBaseImagesList(c) })
	ginEngine.GET("/api/v1/deployments/build", func(c *gin.Context) { buildDeployment(c) })
	ginEngine.GET("/api/v1/deployments/list", func(c *gin.Context) { renderDeploymentsList(c) })
}

func buildBaseImages(c *gin.Context) {
	tag := c.Query("tag")
	if tag == "" {
		c.String(400, "Tag is required for Web API calls")
		return
	}
	c.String(200, "Build process started")
	go func(tag string, forceRebuild bool) {
		dockerbuild.BuildBaseImages(registryBasePath, tag, forceRebuild, true)
	}(tag, c.Query("force-rebuild") != "")
}

func buildDeployment(c *gin.Context) {
	tag := c.Query("tag")
	if tag == "" {
		c.String(400, "Tag is required for Web API calls")
	}
	c.String(200, "Build process started")
	go func(deploymentName string, tag string, deploymentTag string) {
		dockerbuild.BuildDeployment(registryBasePath, deploymentName, tag, deploymentTag, true)
	}(c.Query("name"), tag, c.Query("deployment-tag"))
}

func renderBaseImagesList(c *gin.Context) {
	buildableImages, orphanedImages := dockerbuild.GetBaseImageHeirarchy()

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
		output += "Orphaned Images:\n"
		for _, oi := range orphanedImages {
			output += oi.Name + " (requested from " + oi.ParentName + ")\n"
		}
		c.String(200, output)
	}
}

func renderDeploymentsList(c *gin.Context) {
	deployments := dockerbuild.GetDeployments()

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
		output := ""
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
