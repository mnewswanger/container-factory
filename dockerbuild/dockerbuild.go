package dockerbuild

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"regexp"

	"gitlab.home.mikenewswanger.com/golang/executil"
	"gitlab.home.mikenewswanger.com/golang/filesystem"

	"sync"

	"os/user"

	"os"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
)

// DockerBuild provides build services for docker images
type DockerBuild struct {
	Verbosity           uint8
	Debug               bool
	DockerfileDirectory string
	InternalImagePrefix string
	Tag                 string

	dockerfiles         []dockerfile
	dockerfileHeirarchy map[string][]dockerfile

	imageBoolMap map[string]bool
}

type dockerfile struct {
	name                    string
	parentName              string
	fileName                string
	hasInternalDependencies bool
}

func (db *DockerBuild) initialize() {
	if db.InternalImagePrefix == "" {
		color.Red("Registry Base Path must be specified")
		os.Exit(100)
	}
	db.DockerfileDirectory, _ = homedir.Expand(db.DockerfileDirectory)
	db.DockerfileDirectory = filesystem.ForceTrailingSlash(db.DockerfileDirectory)

	db.dockerfileHeirarchy = make(map[string][]dockerfile)

	if db.Tag == "" {
		currentUser, err := user.Current()
		if err != nil {
			panic(err)
		}
		db.Tag = currentUser.Username
	}
}

// BuildImages builds all docker images by heirarchy
func (db *DockerBuild) BuildImages(forceRebuild bool) {
	db.initialize()

	db.loadDockerImageHeirarchy()

	var waitGroup = sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		db.buildImagesWithChildren("", forceRebuild)
	}()
	waitGroup.Wait()
}

// PrintImageHeirarchy prints the heirachy of dockerfiles to be built to stdout
func (db *DockerBuild) PrintImageHeirarchy() {
	db.initialize()
	db.loadDockerImageHeirarchy()
	db.imageBoolMap = make(map[string]bool)

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

func (db *DockerBuild) buildDockerImageHeirarchy() {
	for _, df := range db.dockerfiles {
		if df.hasInternalDependencies {
			db.dockerfileHeirarchy[df.parentName] = append(db.dockerfileHeirarchy[df.parentName], df)
		} else {
			db.dockerfileHeirarchy[""] = append(db.dockerfileHeirarchy[df.parentName], df)
		}
	}
}

func (db *DockerBuild) buildImagesWithChildren(parent string, forceRebuild bool) {
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
				WorkingDirectory: db.DockerfileDirectory + "/..",
				Debug:            db.Debug,
				Verbosity:        db.Verbosity,
			}.RunWithRealtimeOutput()

			waitGroup.Add(1)
			go func(image string) {
				db.pushImageToRegistry(image)
				waitGroup.Done()
			}(c.name + ":" + db.Tag)

			waitGroup.Add(1)

			// Build all of the children
			go func(parent string) {
				defer waitGroup.Done()
				db.buildImagesWithChildren(parent, forceRebuild)
			}(c.name)
		}
	}
	waitGroup.Wait()
}

func (db *DockerBuild) loadDockerImageHeirarchy() {
	db.loadDockerfiles("")
	db.buildDockerImageHeirarchy()
}

func (db *DockerBuild) loadDockerfiles(subpath string) {
	for _, f := range filesystem.GetDirectoryContents(db.DockerfileDirectory + subpath) {
		if filesystem.IsDirectory(db.DockerfileDirectory + f) {
			db.loadDockerfiles(f + "/")
		} else {
			var role = subpath + f
			var fileName = db.DockerfileDirectory + role
			firstLine, err := ioutil.ReadFile(fileName)
			if err != nil {
				panic(err)
			}
			readbuffer := bytes.NewBuffer(firstLine)
			reader := bufio.NewReader(readbuffer)
			line, _, err := reader.ReadLine()
			if err == nil {
				fromRegex, _ := regexp.Compile("^FROM\\s+((" + db.InternalImagePrefix + ")?(.+))$")
				// matches[1] == image; matches[2] w/ length > 0 == internal; matches[3] == role
				matches := fromRegex.FindStringSubmatch(string(line))

				db.dockerfiles = append(db.dockerfiles, dockerfile{
					name:                    db.InternalImagePrefix + role,
					fileName:                fileName,
					parentName:              matches[1],
					hasInternalDependencies: len(matches[2]) > 0,
				})
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

func (db *DockerBuild) pushImageToRegistry(image string) {
	executil.Command{
		Name:       "Pushing Docker Image to Registry:" + image,
		Executable: "docker",
		Arguments:  []string{"push", image},
		Debug:      db.Debug,
		Verbosity:  db.Verbosity,
	}.RunWithRealtimeOutput()
}
