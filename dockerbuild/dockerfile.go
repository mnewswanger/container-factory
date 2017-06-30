package dockerbuild

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/sirupsen/logrus"

	"go.mikenewswanger.com/utilities/filesystem"
)

func createDynamicDockerfile(targetDirectory string, sourceFilename string, registryBasePath string, tag string) string {
	// Determine the new filename
	h := sha256.New()
	h.Write([]byte(sourceFilename))
	dynamicDockerfileFilename := targetDirectory + "/" + hex.EncodeToString(h.Sum(nil))

	logger.WithFields(logrus.Fields{
		"temp_dir":        targetDirectory,
		"source_filename": sourceFilename,
		"target_filename": dynamicDockerfileFilename,
	})

	fileContents, err := filesystem.LoadFileString(sourceFilename)
	if err != nil {
		panic(err)
	}

	var matches = fromSplitRegex.FindAllStringSubmatch(fileContents, -1)

	for _, match := range matches {
		if len(match[2]) > 0 {
			fileContents = strings.Replace(fileContents, match[1], filesystem.ForceTrailingSlash(registryBasePath)+match[3]+":"+tag, 1)
		}
	}

	// Write out the new file with the tagged base image
	if err = filesystem.WriteFile(
		dynamicDockerfileFilename,
		[]byte(fileContents),
		0644); err != nil {
		panic(err)
	}

	return dynamicDockerfileFilename
}
