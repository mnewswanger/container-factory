package dockerbuild

import (
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"

	"go.mikenewswanger.com/utilities/executil"
)

// BuildDeployment builds a docker image for a code deployment
func (db *DockerBuild) BuildDeployment(deploymentName string, pushToRemote bool) {
	db.initialize()

	if db.DockerRegistryBasePath == "" {
		db.Logger.Fatal("Registry Base Path must be specified")
		os.Exit(100)
	}

	var deploymentFilename = db.deploymentDirectory + deploymentName
	if !db.fs.IsFile(deploymentFilename) {
		color.Red("Deployment does not exist: " + deploymentName)
		os.Exit(101)
	}
	var tempDir, _ = ioutil.TempDir(db.deploymentDirectory, ".tmp-")
	var dockerfile = db.createDynamicDockerfile(tempDir+"/", deploymentFilename)

	var imageName = db.DockerRegistryBasePath + "/deployments/" + deploymentName + ":" + db.DeploymentTag
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
		WorkingDirectory: db.deploymentDirectory,
		Verbosity:        db.Verbosity,
	}
	if err := cmd.Run(); err == nil {
		if err := db.pushImageToRegistry(imageName); err != nil {
			logrus.Error("Failed to push image to remote registry")
		}
	} else {
		logrus.Error("Deployment failed to build")
	}

	db.fs.RemoveDirectory(tempDir, true)
}

// GetDeployments prints a list of configured deployments
func (db *DockerBuild) GetDeployments() []string {
	db.initialize()

	return db.getFolderDeployments("")
}

func (db *DockerBuild) getFolderDeployments(subpath string) []string {
	var deployments = []string{}
	var directoryContents []string
	var err error
	directoryContents, err = db.fs.GetDirectoryContents(db.deploymentDirectory + subpath)
	if err != nil {
		panic(err)
	}

	for _, f := range directoryContents {
		var relativeFile = subpath + f

		// Skip hidden files
		if string(f[0]) == "." {
			continue
		}

		// Loop through children; iterate any subfolders
		if db.fs.IsFile(db.deploymentDirectory + relativeFile) {
			deployments = append(deployments, relativeFile)
		} else {
			deployments = append(deployments, db.getFolderDeployments(relativeFile+"/")...)
		}
	}

	return deployments
}
