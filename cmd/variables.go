package cmd

type flags struct {
	verbosity              int
	deploymentImageTag     string
	dockerBaseDirectory    string
	dockerRegistryBasePath string
	forceRebuild           bool
	imageTag               string
	localOnly              bool
	outputFormat           string
}

var commandLineFlags = flags{}
