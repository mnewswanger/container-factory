package dockerbuild

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"go.mikenewswanger.com/utilities/executil"
	"go.mikenewswanger.com/utilities/filesystem"
)

// BuildBaseImages builds all docker images by heirarchy
func BuildBaseImages(dockerRegistryBasePath string, tag string, forceRebuild bool, pushToRemote bool) {
	tag = getDefaultTag(tag)

	logger.WithFields(logrus.Fields{
		"tag": tag,
	}).Info("Building all images")

	if forceRebuild {
		logger.Warn("Forcing a rebuild.  Caches will not be used.")
	}
	if !pushToRemote {
		logger.Warn("Push to remote is disabled")
	}

	tempDir, _ := ioutil.TempDir(dockerfileDirectory, ".tmp-")
	defer filesystem.RemoveDirectory(tempDir, true)
	logger.WithFields(logrus.Fields{
		"path": tempDir,
	}).Debug("Created temp directory")

	var waitGroup = sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		buildBaseImagesWithChildren(tempDir, dockerRegistryBasePath, tag, "", forceRebuild, pushToRemote)
	}()
	waitGroup.Wait()
}

// GetBaseImageHeirarchy prints the heirachy of dockerfiles to be built to stdout
// Returns buildable images and orphaned images respectively
func GetBaseImageHeirarchy() ([]DockerBuildableImage, []DockerOrphanedImage) {
	return buildableImages, orphanedImages
}

func buildBaseImagesWithChildren(tempDir string, registryBasePath string, tag string, parent string, forceRebuild bool, pushToRemote bool) {
	var waitGroup = sync.WaitGroup{}
	children, hasChildren := dockerfileHeirarchy[parent]
	if hasChildren {
		for _, c := range children {
			var imageName = filesystem.ForceTrailingSlash(registryBasePath) + c.name
			logger.WithFields(logrus.Fields{
				"docker_image": imageName,
			}).Info("Building Image")

			arguments := []string{"build", "-t", imageName + ":" + tag, "-f", createDynamicDockerfile(tempDir, c.filename, registryBasePath, tag)}
			if forceRebuild {
				arguments = append(arguments, "--no-cache=true")
			}
			arguments = append(arguments, ".")
			var cmd = executil.Command{
				Name:             "Building Docker Image: " + imageName + ":" + tag,
				Executable:       "docker",
				Arguments:        arguments,
				WorkingDirectory: dockerBaseDirectory,
			}
			if err := cmd.Run(); err == nil {
				waitGroup.Add(1)
				if pushToRemote {
					waitGroup.Add(1)
					go func(image string) {
						err := pushImageToRegistry(image)
						if err != nil {
							logger.WithFields(logrus.Fields{
								"docker_image": imageName,
							}).Error(err)
						}
						waitGroup.Done()
					}(imageName + ":" + tag)
				}

				// Build all of the children
				go func(parent string) {
					defer waitGroup.Done()
					buildBaseImagesWithChildren(tempDir, registryBasePath, tag, parent, forceRebuild, pushToRemote)
				}(c.name)
			} else {
				logger.WithFields(logrus.Fields{
					"docker_image": imageName,
				}).Error("Image failed to build")
			}
		}
	}
	waitGroup.Wait()
}

func buildDockerImageHeirarchy() (map[string][]*dockerfile, []DockerBuildableImage, []DockerOrphanedImage) {
	logger.Info("Building Docker image heirarchy")
	dfh := map[string][]*dockerfile{}
	allImages := loadBaseImageDockerfiles("")
	for _, df := range allImages {
		if df.hasInternalDependencies {
			dfh[df.parentName] = append(dfh[df.parentName], df)
		} else {
			dfh[""] = append(dfh[""], df)
		}
	}

	// Loop buildable to determine all images that are buildable
	bi := getChildImages(dfh, "")
	oi := []DockerOrphanedImage{}
	for _, df := range allImages {
		if !df.isBuildable {
			oi = append(oi, DockerOrphanedImage{
				Name:       df.name,
				ParentName: df.parentName,
			})
		}
	}

	return dfh, bi, oi
}

func getChildImages(dfh map[string][]*dockerfile, parent string) []DockerBuildableImage {
	imageHeirarchy := []DockerBuildableImage{}
	children, hasChildren := dfh[parent]
	if hasChildren {
		for _, df := range children {
			df.isBuildable = true
			imageHeirarchy = append(imageHeirarchy, DockerBuildableImage{
				Name:     df.name,
				Children: getChildImages(dfh, df.name),
			})
		}
	}
	return imageHeirarchy
}

// loadBaseImageDockerfiles loads base image dockerfiles from a directory recursively
func loadBaseImageDockerfiles(subpath string) map[string]*dockerfile {
	logger.WithFields(logrus.Fields{
		"namespace": "/" + subpath,
	}).Debug("Processing DockerFiles")

	dockerfiles := map[string]*dockerfile{}

	directoryContents, err := filesystem.GetDirectoryContents(dockerfileDirectory + subpath)
	if err != nil {
		logger.Panic(err)
	}
	for _, f := range directoryContents {
		var relativeFile = subpath + f

		// Skip hidden files
		if string(f[0]) == "." {
			continue
		}

		// Loop through children; iterate any subfolders
		if filesystem.IsDirectory(dockerfileDirectory + relativeFile) {
			for n, df := range loadBaseImageDockerfiles(relativeFile + "/") {
				dockerfiles[n] = df
			}
		} else {
			role := relativeFile
			fileName := dockerfileDirectory + role
			firstLine, err := ioutil.ReadFile(fileName)
			if err != nil {
				panic(err)
			}
			readbuffer := bytes.NewBuffer(firstLine)
			reader := bufio.NewReader(readbuffer)
			line, _, err := reader.ReadLine()
			if err == nil {
				matches := fromSplitRegex.FindStringSubmatch(string(line))

				parentName := ""
				hasInternalDependencies := len(matches[2]) > 0

				if hasInternalDependencies {
					parentName = strings.Replace(matches[1], matches[2], "", 1)
				}

				var df = dockerfile{
					name:                    role,
					filename:                fileName,
					parentName:              parentName,
					hasInternalDependencies: hasInternalDependencies,
				}
				dockerfiles[df.name] = &df
			}
		}
	}
	return dockerfiles
}
