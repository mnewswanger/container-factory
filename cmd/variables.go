package cmd

type flags struct {
	verbosity              int
	debug                  bool
	deploymentImageTag     string
	dockerBaseDirectory    string
	forceRebuild           bool
	imageTag               string
	dockerRegistryBasePath string
	localOnly              bool
}

var commandLineFlags = flags{}
