package cmd

type flags struct {
	verbosity              int
	debug                  bool
	dockerfileDirectory    string
	forceRebuild           bool
	imageTag               string
	dockerRegistryBasePath string
}

var commandLineFlags = flags{}
