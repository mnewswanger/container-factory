package dockerbuild

import (
	"os/user"
	"regexp"

	"github.com/sirupsen/logrus"

	"go.mikenewswanger.com/utilities/filesystem"
)

// DockerBuild provides build services for docker images
type DockerBuild struct {
	Verbosity              uint8
	DockerBaseDirectory    string
	DockerRegistryBasePath string
	DeploymentTag          string
	Logger                 *logrus.Logger
	Tag                    string

	deploymentDirectory string
	dockerfileDirectory string
	dockerfiles         map[string]*dockerfile
	dockerfileHeirarchy map[string][]*dockerfile
	fs                  *filesystem.Filesystem
	initialized         bool

	fromSplitRegex *regexp.Regexp
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
}

func (db *DockerBuild) initialize() {
	// No need to re-initialize if we've alredy done it
	if db.initialized {
		return
	}

	if db.Logger == nil {
		// Set up the logger
		db.Logger = logrus.New()
		switch db.Verbosity {
		case 0:
			db.Logger.Level = logrus.ErrorLevel
			break
		case 1:
			db.Logger.Level = logrus.WarnLevel
			break
		case 2:
			fallthrough
		case 3:
			db.Logger.Level = logrus.InfoLevel
			break
		default:
			db.Logger.Level = logrus.DebugLevel
			break
		}
	}

	db.Logger.Info("Initializing DockerBuild System")

	if db.DockerRegistryBasePath == "" {
		db.Logger.Fatal("Registry Base Path must be specified")
	}

	// matches[1] => image; matches[2] w/ length > 0 => internal; matches[3] => role
	db.fromSplitRegex, _ = regexp.Compile("FROM\\s+(({{\\s+local\\s+}}/)?([\\w\\-\\_\\/\\:\\.\\{\\}]+))([\\s\\n])?")

	db.Logger.Debug("Initializing Filesystem utility")
	db.fs = &filesystem.Filesystem{
		Verbosity: db.Verbosity,
		Logger:    db.Logger,
	}

	// Determine environment paths
	db.dockerfileDirectory, _ = db.fs.BuildAbsolutePathFromHome(db.DockerBaseDirectory + "/dockerfiles/")
	db.Logger.WithFields(logrus.Fields{
		"path": db.dockerfileDirectory,
	}).Debug("Set dockerfile directory")
	db.deploymentDirectory, _ = db.fs.BuildAbsolutePathFromHome(db.DockerBaseDirectory + "/deployments/")
	db.Logger.WithFields(logrus.Fields{
		"path": db.deploymentDirectory,
	}).Debug("Set deployment directory")

	db.dockerfiles = make(map[string]*dockerfile)

	db.setTags()

	db.initialized = true
}

func (db *DockerBuild) setTags() {
	// Determine tagging for images and deployments
	if db.Tag == "" {
		currentUser, err := user.Current()
		if err != nil {
			panic(err)
		}
		db.Tag = currentUser.Username
	}
	if db.DeploymentTag == "" {
		db.DeploymentTag = db.Tag
	}
	db.Logger.WithFields(logrus.Fields{
		"tag":            db.Tag,
		"deployment_tag": db.DeploymentTag,
	}).Info("Set build tags")
}
