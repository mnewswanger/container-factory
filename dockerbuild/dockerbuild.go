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
	"github.com/mitchellh/go-homedir"

	"gitlab.home.mikenewswanger.com/golang/executil"
	"gitlab.home.mikenewswanger.com/golang/filesystem"
)

// DockerBuild provides build services for docker images
type DockerBuild struct {
	Verbosity              uint8
	Debug                  bool
	DockerBaseDirectory    string
	DockerRegistryBasePath string
	Tag                    string

	deploymentDirectory string
	dockerfileDirectory string
	dockerfiles         []*dockerfile
	dockerfileHeirarchy map[string][]*dockerfile
	imageBoolMap        map[string]bool
	internalImagePrefix string

	fromSplitRegex *regexp.Regexp
}

type dockerfile struct {
	name                    string
	parentName              string
	fileName                string
	hasInternalDependencies bool
}

func (db *DockerBuild) initialize() {
	if db.DockerRegistryBasePath == "" {
		color.Red("Registry Base Path must be specified")
		os.Exit(100)
	}
	var imageBaseRegex, _ = regexp.Compile("^([\\w\\.\\-\\_]+)(:\\d+)?(/.*)")

	// matches[1] => host; matches[2] => port; matches[3] => image
	var matches = imageBaseRegex.FindStringSubmatch(db.DockerRegistryBasePath)
	db.internalImagePrefix = matches[1] + matches[3]

	// matches[1] => image; matches[2] w/ length > 0 => internal; matches[3] => role
	db.fromSplitRegex, _ = regexp.Compile("FROM\\s+((" + db.internalImagePrefix + ")?([\\w\\-\\_\\/\\:\\.]+))\\s*\\n?")

	db.dockerfileDirectory, _ = homedir.Expand(filesystem.ForceTrailingSlash(db.DockerBaseDirectory) + "dockerfiles/")
	db.deploymentDirectory, _ = homedir.Expand(filesystem.ForceTrailingSlash(db.DockerBaseDirectory) + "deployments/")

	db.dockerfileHeirarchy = make(map[string][]*dockerfile)

	if db.Tag == "" {
		currentUser, err := user.Current()
		if err != nil {
			panic(err)
		}
		db.Tag = currentUser.Username
	}
}

// BuildBaseImages builds all docker images by heirarchy
func (db *DockerBuild) BuildBaseImages(forceRebuild bool, pushToRemote bool) {
	db.initialize()

	var tempDir = db.dockerfileDirectory + ".tmp/"

	db.loadDockerImageHeirarchy()
	db.createDynamicBuildFiles(tempDir)

	var waitGroup = sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		db.buildImagesWithChildren("", forceRebuild, pushToRemote)
	}()
	waitGroup.Wait()

	filesystem.RemoveDirectory(tempDir, true)
}

// BuildDeployment builds a docker image for a code deployment
func (db *DockerBuild) BuildDeployment(deploymentName string, pushToRemote bool) {
	var tempDir = db.deploymentDirectory + ".tmp"
	filesystem.RemoveDirectory(tempDir, true)
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
			var arguments = []string{"build", "-t", c.name + ":" + db.Tag, "-f", c.fileName}
			if forceRebuild {
				arguments = append(arguments, "--no-cache=true")
			}
			arguments = append(arguments, ".")
			executil.Command{
				Name:             "Building Docker Image: " + c.name + ":" + db.Tag,
				Executable:       "docker",
				Arguments:        arguments,
				WorkingDirectory: db.dockerfileDirectory + "/..",
				Debug:            db.Debug,
				Verbosity:        db.Verbosity,
			}.RunWithRealtimeOutput()
			waitGroup.Add(1)
			if pushToRemote {
				go func(image string) {
					db.pushImageToRegistry(image)
					waitGroup.Done()
				}(c.name + ":" + db.Tag)
				waitGroup.Add(1)
			}

			// Build all of the children
			go func(parent string) {
				defer waitGroup.Done()
				db.buildImagesWithChildren(parent, forceRebuild, pushToRemote)
			}(c.name)
		}
	}
	waitGroup.Wait()
}

func (db *DockerBuild) createDynamicBuildFiles(tempDir string) {
	filesystem.CreateDirectory(tempDir)

	for i := range db.dockerfiles {
		db.dockerfiles[i].fileName = db.createDynamicDockerfile(tempDir, db.dockerfiles[i].fileName)
	}
}

func (db *DockerBuild) createDynamicDockerfile(tempDir string, sourceFilename string) string {
	// Determine the new filename
	var h = sha256.New()
	h.Write([]byte(sourceFilename))
	var dynamicDockerfileFilename = tempDir + hex.EncodeToString(h.Sum(nil))

	var fileContents = filesystem.LoadFileIfExists(sourceFilename)

	var matches = db.fromSplitRegex.FindAllStringSubmatch(fileContents, -1)

	for _, match := range matches {
		if len(match[2]) > 0 {
			fileContents = strings.Replace(fileContents, match[1], db.DockerRegistryBasePath+match[3]+":"+db.Tag, 1)
		}
	}

	// Write out the new file with the tagged base image
	filesystem.WriteFile(
		dynamicDockerfileFilename,
		[]byte(fileContents),
		0644)

	return dynamicDockerfileFilename
}

func (db *DockerBuild) loadDockerImageHeirarchy() {
	db.loadDockerfiles("")
	db.buildDockerImageHeirarchy()
}

func (db *DockerBuild) loadDockerfiles(subpath string) {
	for _, f := range filesystem.GetDirectoryContents(db.dockerfileDirectory + subpath) {
		var relativeFile = subpath + f

		// Skip hidden files
		if string(f[0]) == "." {
			continue
		}

		// Loop through children; iterate any subfolders
		if filesystem.IsDirectory(db.dockerfileDirectory + relativeFile) {
			db.loadDockerfiles(relativeFile + "/")
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

				var df = dockerfile{
					name:                    db.internalImagePrefix + role,
					fileName:                fileName,
					parentName:              matches[1],
					hasInternalDependencies: len(matches[2]) > 0,
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
	for _, f := range filesystem.GetDirectoryContents(db.deploymentDirectory + subpath) {
		var relativeFile = subpath + f

		// Skip hidden files
		if string(f[0]) == "." {
			continue
		}

		// Loop through children; iterate any subfolders
		if filesystem.IsFile(db.deploymentDirectory + relativeFile) {
			println(relativeFile)
		} else {
			db.printFolderDeployments(relativeFile + "/")
		}
	}
}

func (db *DockerBuild) pushImageToRegistry(image string) {
	executil.Command{
		Name:       "Pushing Docker Image to Registry:" + image,
		Executable: "docker",
		Arguments:  []string{"push", image},
		Debug:      db.Debug,
		Verbosity:  db.Verbosity,
	}.RunWithRealtimeOutput()
}
