package dockerbuild

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strings"
	"sync"

	"gitlab.home.mikenewswanger.com/golang/executil"

	"github.com/sirupsen/logrus"
)

// BuildBaseImages builds all docker images by heirarchy
func (db *DockerBuild) BuildBaseImages(forceRebuild bool, pushToRemote bool) {
	db.initialize()

	db.createBaseImageHeirarchy()

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
		db.buildBaseImagesWithChildren("", forceRebuild, pushToRemote)
	}()
	waitGroup.Wait()

	db.fs.RemoveDirectory(tempDir, true)
}

// GetBaseImageHeirarchy prints the heirachy of dockerfiles to be built to stdout
// Returns buildable images and orphaned images respectively
func (db *DockerBuild) GetBaseImageHeirarchy() ([]DockerBuildableImage, []DockerOrphanedImage) {
	db.initialize()
	db.createBaseImageHeirarchy()

	var allImages = make(map[string]bool)
	for _, df := range db.dockerfiles {
		allImages[df.name] = false
	}

	var buildableImages = db.getChildImages("", allImages)

	var orphanedImages = []DockerOrphanedImage{}
	for imageName, buildable := range allImages {
		if !buildable {
			orphanedImages = append(orphanedImages, DockerOrphanedImage{
				Name:       imageName,
				ParentName: db.dockerfiles[imageName].parentName,
			})
		}
	}

	return buildableImages, orphanedImages
}

func (db *DockerBuild) buildBaseImagesWithChildren(parent string, forceRebuild bool, pushToRemote bool) {
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
					waitGroup.Add(1)
					go func(image string) {
						var err = db.pushImageToRegistry(image)
						if err != nil {
							db.Logger.WithFields(logrus.Fields{
								"docker_image": imageName,
							}).Error(err)
						}
						waitGroup.Done()
					}(imageName + ":" + db.Tag)
				}

				// Build all of the children
				go func(parent string) {
					defer waitGroup.Done()
					db.buildBaseImagesWithChildren(parent, forceRebuild, pushToRemote)
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

func (db *DockerBuild) createBaseImageHeirarchy() {
	db.Logger.Info("Loading DockerFiles to build")
	db.loadBaseImageDockerfiles("")
	db.buildDockerImageHeirarchy()
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

func (db *DockerBuild) getChildImages(parent string, buildableImages map[string]bool) []DockerBuildableImage {
	var imageHeirarchy = []DockerBuildableImage{}
	children, hasChildren := db.dockerfileHeirarchy[parent]
	if hasChildren {
		var image DockerBuildableImage
		for _, df := range children {
			image = DockerBuildableImage{
				Name:     df.name,
				Children: db.getChildImages(df.name, buildableImages),
			}
			imageHeirarchy = append(imageHeirarchy, image)
			buildableImages[df.name] = true
		}
	}
	return imageHeirarchy
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
				db.dockerfiles[df.name] = &df
			}
		}
	}
}
