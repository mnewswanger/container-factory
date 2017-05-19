package cmd

import "github.com/sirupsen/logrus"

type flags struct {
	verbosity              int
	deploymentImageTag     string
	dockerBaseDirectory    string
	dockerRegistryBasePath string
	forceRebuild           bool
	imageTag               string
	listenPort             uint16
	localOnly              bool
	outputFormat           string
}

var commandLineFlags = flags{}

var logger = logrus.New()
