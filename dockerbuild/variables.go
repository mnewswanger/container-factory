package dockerbuild

import (
	"regexp"

	"github.com/sirupsen/logrus"
	"go.mikenewswanger.com/utilities/executil"
	"go.mikenewswanger.com/utilities/filesystem"
)

var dockerBaseDirectory string
var dockerfileDirectory string
var deploymentDirectory string
var logger = logrus.New()
var verbosity = uint8(0)
var deployments = []string{}

// Populated during BuildInventory()
var buildableImages []DockerBuildableImage
var dockerfileHeirarchy map[string][]*dockerfile
var orphanedImages []DockerOrphanedImage

// matches[1] => image; matches[2] w/ length > 0 => internal; matches[3] => role
var fromSplitRegex, _ = regexp.Compile("FROM\\s+(({{\\s+local\\s+}}/)?([\\w\\-\\_\\/\\:\\.\\{\\}]+))([\\s\\n])?")

// SetDockerBaseDirectory sets the base directory to use for the docker build process and caches inventory into memory
func SetDockerBaseDirectory(path string) {
	if path == "" {
		logger.Fatal("Registry Base Path must be specified")
	}

	var err error
	dockerBaseDirectory, err = filesystem.BuildAbsolutePathFromHome(path)
	logger.WithFields(logrus.Fields{
		"docker_base_directory": dockerBaseDirectory,
	}).Info("Setting Docker base directory")
	if err != nil {
		logger.Error(err)
	}
	// Determine environment paths
	dockerfileDirectory, _ = filesystem.BuildAbsolutePathFromHome(dockerBaseDirectory + "/dockerfiles/")
	logger.WithFields(logrus.Fields{
		"path": dockerfileDirectory,
	}).Debug("Set dockerfile directory")
	deploymentDirectory, _ = filesystem.BuildAbsolutePathFromHome(dockerBaseDirectory + "/deployments/")
	logger.WithFields(logrus.Fields{
		"path": deploymentDirectory,
	}).Debug("Set deployment directory")

	BuildInventory()
}

// SetLogger allows overriding the default logger
func SetLogger(l *logrus.Logger) {
	logger = l
	executil.SetLogger(l)
	filesystem.SetLogger(l)
}

// SetVerbosity allows the caller to increase verbosity of the package
func SetVerbosity(v uint8) {
	verbosity = v
	executil.SetVerbosity(v)
	filesystem.SetVerbosity(v)
}
