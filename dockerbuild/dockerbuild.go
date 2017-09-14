package dockerbuild

import "os/user"
import "strings"

// DockerBuild provides build services for docker images
type DockerBuild struct {
}

// DockerBuildableImage provides a structure to export the docker image heirarchy
type DockerBuildableImage struct {
	Name     string                 `json:"image_name"`
	Children []DockerBuildableImage `json:"children"`
}

// DockerOrphanedImage provides a structure to export docker images that are not buildable
type DockerOrphanedImage struct {
	Name       string `json:"image_name"`
	ParentName string `json:"parent_image_name"`
}

type dockerfile struct {
	name                    string
	parentName              string
	filename                string
	hasInternalDependencies bool
	isBuildable             bool
}

// BuildInventory loads available base images and deployments into memory
func BuildInventory() {
	deployments = getFolderDeployments("")
	dockerfileHeirarchy, buildableImages, orphanedImages = buildDockerImageHeirarchy()
}

func getDefaultTag(tag string) string {
	if tag == "" {
		currentUser, err := user.Current()
		if err != nil {
			logger.Fatal(err)
		}
		tag = currentUser.Username
	}
	return tag
}

func isValidDockerfile(filename string) bool {
	if string(filename[0]) == "." || strings.ToLower(filename) == "readme.md" {
		return false
	}
	return true
}
