package dockerbuild

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"

	"go.mikenewswanger.com/utilities/executil"
	"go.mikenewswanger.com/utilities/filesystem"
)

// BuildDeployment builds a docker image for a code deployment
func BuildDeployment(registryBasePath string, deploymentName string, buildTargetTag string, deploymentTag string, pushToRemote bool) {
	if registryBasePath == "" {
		logger.Panic("Registry Base Path must be specified")
	}
	buildTargetTag = getDefaultTag(buildTargetTag)
	if deploymentTag == "" {
		deploymentTag = buildTargetTag
	}

	deploymentFilename := deploymentDirectory + deploymentName
	if !filesystem.IsFile(deploymentFilename) {
		logger.Panic("Deployment does not exist: " + deploymentName)
	}
	tempDir, _ := ioutil.TempDir(deploymentDirectory, ".tmp-")
	dockerfile := createDynamicDockerfile(tempDir+"/", deploymentFilename, registryBasePath, buildTargetTag)

	var imageName = registryBasePath + "/deployments/" + deploymentName + ":" + deploymentTag
	var cmd = executil.Command{
		Name:       "Build Deployment - " + imageName,
		Executable: "docker",
		Arguments: []string{
			"build",
			"--no-cache",
			"-t",
			imageName,
			"-f",
			dockerfile,
			".",
		},
		WorkingDirectory: dockerBaseDirectory,
	}
	if err := cmd.Run(); err == nil {
		if pushToRemote {
			if err := pushImageToRegistry(imageName); err != nil {
				logrus.Error("Failed to push image to remote registry")
			}
		}
	} else {
		logrus.Error("Deployment failed to build")
	}

	filesystem.RemoveDirectory(tempDir, true)
}

// GetDeployments prints a list of configured deployments
func GetDeployments() []string {
	return deployments
}

func getFolderDeployments(subpath string) []string {
	d := []string{}
	directoryContents, err := filesystem.GetDirectoryContents(deploymentDirectory + subpath)
	if err != nil {
		logger.Panic(err)
	}

	for _, f := range directoryContents {
		relativeFile := subpath + f

		// Skip hidden files
		if string(f[0]) == "." {
			continue
		}

		// Loop through children; iterate any subfolders
		if filesystem.IsFile(deploymentDirectory + relativeFile) {
			d = append(d, relativeFile)
		} else {
			d = append(d, getFolderDeployments(relativeFile+"/")...)
		}
	}

	return d
}
