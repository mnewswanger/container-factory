package dockerbuild

import (
	"github.com/sirupsen/logrus"

	"go.mikenewswanger.com/utilities/executil"
)

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
		}

		db.Logger.WithFields(logrus.Fields{
			"docker_image":      image,
			"retries_remaining": retries,
		}).Warn("Failed to push image to registry")
	}

	return err
}
