package dockerbuild

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"os/user"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"

	"gitlab.home.mikenewswanger.com/golang/executil"
	"gitlab.home.mikenewswanger.com/golang/filesystem"
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
	dockerfiles         []*dockerfile
	dockerfileHeirarchy map[string][]*dockerfile
	fs                  *filesystem.Filesystem
	imageBoolMap        map[string]bool
	initialized         bool

	fromSplitRegex *regexp.Regexp
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

	db.initialized = true
}

// BuildBaseImages builds all docker images by heirarchy
func (db *DockerBuild) BuildBaseImages(forceRebuild bool, pushToRemote bool) {
	db.initialize()

	db.loadDockerImageHeirarchy()

	db.Logger.Info("Building all images")

	if forceRebuild {
		db.Logger.Warn("Forcing a rebuild.  Caches will not be used.")
	}
	if !pushToRemote {
		db.Logger.Warn("Push to remote is disabled")
	}

	var tempDir, _ = ioutil.TempDir(db.dockerfileDirectory, ".tmp-")
	db.Logger.WithFields(logrus.Fields{
		"path": tempDir,
	}).Debug("Created temp directory")
	db.createDynamicBuildFiles(tempDir + "/")

	var waitGroup = sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		db.buildImagesWithChildren("", forceRebuild, pushToRemote)
	}()
	waitGroup.Wait()

	db.fs.RemoveDirectory(tempDir, true)
}

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

	var imageName = db.DockerRegistryBasePath + "deployments/" + deploymentName + ":" + db.DeploymentTag
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

// PrintBaseImageHeirarchy prints the heirachy of dockerfiles to be built to stdout
func (db *DockerBuild) PrintBaseImageHeirarchy() {
	db.initialize()
	db.loadDockerImageHeirarchy()
	db.imageBoolMap = make(map[string]bool)

	for _, df := range db.dockerfiles {
		db.imageBoolMap[df.name] = false
	}

	println("Buildable Images:")
	db.printChildImages("", 0)
	println("")

	println("Orphaned Images:")
	for dfname, printed := range db.imageBoolMap {
		if !printed {
			color.Red(dfname)
		}
	}
}

// PrintDeployments prints a list of configured deployments
func (db *DockerBuild) PrintDeployments() {
	db.initialize()

	db.printFolderDeployments("")
	println()
}

func (db *DockerBuild) buildDockerImageHeirarchy() {
	db.Logger.Info("Building Docker image heirarchy")
	db.dockerfileHeirarchy = make(map[string][]*dockerfile)
	for _, df := range db.dockerfiles {
		if df.hasInternalDependencies {
			db.dockerfileHeirarchy[df.parentName] = append(db.dockerfileHeirarchy[df.parentName], df)
		} else {
			db.dockerfileHeirarchy[""] = append(db.dockerfileHeirarchy[""], df)
		}
	}
}

func (db *DockerBuild) buildImagesWithChildren(parent string, forceRebuild bool, pushToRemote bool) {
	var waitGroup = sync.WaitGroup{}
	children, hasChildren := db.dockerfileHeirarchy[parent]
	if hasChildren {
		for _, c := range children {
			var imageName = db.fs.ForceTrailingSlash(db.DockerRegistryBasePath) + c.name
			db.Logger.WithFields(logrus.Fields{
				"docker_image": imageName,
			}).Info("Building Image")

			var arguments = []string{"build", "-t", imageName + ":" + db.Tag, "-f", c.filename}
			if forceRebuild {
				arguments = append(arguments, "--no-cache=true")
			}
			arguments = append(arguments, ".")
			var cmd = executil.Command{
				Name:             "Building Docker Image: " + imageName + ":" + db.Tag,
				Executable:       "docker",
				Arguments:        arguments,
				WorkingDirectory: db.dockerfileDirectory + "/..",
				Verbosity:        db.Verbosity,
			}
			if err := cmd.Run(); err == nil {
				waitGroup.Add(1)
				if pushToRemote {
					go func(image string) {
						var err = db.pushImageToRegistry(image)
						if err != nil {
							db.Logger.WithFields(logrus.Fields{
								"docker_image": imageName,
							}).Error(err)
						}
						waitGroup.Done()
					}(imageName + ":" + db.Tag)
					waitGroup.Add(1)
				}

				// Build all of the children
				go func(parent string) {
					defer waitGroup.Done()
					db.buildImagesWithChildren(parent, forceRebuild, pushToRemote)
				}(c.name)
			} else {
				db.Logger.WithFields(logrus.Fields{
					"docker_image": imageName,
				}).Error("Image failed to build")
			}
		}
	}
	waitGroup.Wait()
}

func (db *DockerBuild) createDynamicBuildFiles(targetDirectory string) {
	db.Logger.Info("Generating Dockerfiles")
	for _, dockerfile := range db.dockerfiles {
		db.Logger.WithFields(logrus.Fields{
			"name":     dockerfile.name,
			"filename": dockerfile.filename,
		}).Debug("Compiling Dockerfile")
		dockerfile.filename = db.createDynamicDockerfile(targetDirectory, dockerfile.filename)
	}
}

func (db *DockerBuild) createDynamicDockerfile(targetDirectory string, sourceFilename string) string {
	var err error
	var fileContents string

	// Determine the new filename
	var h = sha256.New()
	h.Write([]byte(sourceFilename))
	var dynamicDockerfileFilename = targetDirectory + hex.EncodeToString(h.Sum(nil))

	fileContents, err = db.fs.LoadFileIfExists(sourceFilename)
	if err != nil {
		panic(err)
	}

	var matches = db.fromSplitRegex.FindAllStringSubmatch(fileContents, -1)

	for _, match := range matches {
		if len(match[2]) > 0 {
			fileContents = strings.Replace(fileContents, match[1], db.fs.ForceTrailingSlash(db.DockerRegistryBasePath)+match[3]+":"+db.Tag, 1)
		}
	}

	// Write out the new file with the tagged base image
	if e := db.fs.WriteFile(
		dynamicDockerfileFilename,
		[]byte(fileContents),
		0644); e != nil {
		panic(e)
	}

	return dynamicDockerfileFilename
}

func (db *DockerBuild) loadDockerImageHeirarchy() {
	db.Logger.Info("Loading DockerFiles to build")
	db.loadBaseImageDockerfiles("")
	db.buildDockerImageHeirarchy()
}

func (db *DockerBuild) loadBaseImageDockerfiles(subpath string) {
	db.Logger.WithFields(logrus.Fields{
		"namespace": "/" + subpath,
	}).Debug("Processing DockerFiles")

	var directoryContents []string
	var err error

	directoryContents, err = db.fs.GetDirectoryContents(db.dockerfileDirectory + subpath)
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
		if db.fs.IsDirectory(db.dockerfileDirectory + relativeFile) {
			db.loadBaseImageDockerfiles(relativeFile + "/")
		} else {
			var role = relativeFile
			var fileName = db.dockerfileDirectory + role
			firstLine, err := ioutil.ReadFile(fileName)
			if err != nil {
				panic(err)
			}
			readbuffer := bytes.NewBuffer(firstLine)
			reader := bufio.NewReader(readbuffer)
			line, _, err := reader.ReadLine()
			if err == nil {
				matches := db.fromSplitRegex.FindStringSubmatch(string(line))

				var parentName = ""
				var hasInternalDependencies = len(matches[2]) > 0

				if hasInternalDependencies {
					parentName = strings.Replace(matches[1], matches[2], "", 1)
				}

				var df = dockerfile{
					name:                    role,
					filename:                fileName,
					parentName:              parentName,
					hasInternalDependencies: hasInternalDependencies,
				}
				db.dockerfiles = append(db.dockerfiles, &df)
			}
		}
	}
}

func (db *DockerBuild) printChildImages(parent string, level uint8) {
	children, hasChildren := db.dockerfileHeirarchy[parent]
	if hasChildren {
		for _, df := range children {
			if level > 0 {
				print(color.GreenString("|--"))
			}
			for i := level; i > 1; i-- {
				print(color.GreenString("---"))
			}
			color.Green(df.name)
			db.printChildImages(df.name, level+1)
			db.imageBoolMap[df.name] = true
		}
	}
}

func (db *DockerBuild) printFolderDeployments(subpath string) {
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
			println(relativeFile)
		} else {
			db.printFolderDeployments(relativeFile + "/")
		}
	}
}

func (db *DockerBuild) pushImageToRegistry(image string) error {
	var err error
	var cmd = executil.Command{
		Name:       "Pushing Docker Image to Registry: " + image,
		Executable: "docker",
		Arguments:  []string{"push", image},
		Verbosity:  db.Verbosity,
	}

	for retries := 2; retries >= 0; retries-- {
		err = cmd.Run()
		if err == nil {
			return err
		} else {
			db.Logger.WithFields(logrus.Fields{
				"docker_image":      image,
				"retries_remaining": retries,
			}).Warn("Failed to push image to registry")
		}
	}

	return err
}
