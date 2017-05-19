package dockerbuild

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func (db *DockerBuild) createDynamicDockerfile(targetDirectory string, sourceFilename string) string {
	var err error
	var fileContents string

	// Determine the new filename
	var h = sha256.New()
	h.Write([]byte(sourceFilename))
	var dynamicDockerfileFilename = targetDirectory + hex.EncodeToString(h.Sum(nil))

	fileContents, err = db.fs.LoadFileIfExists(sourceFilename)
	if err != nil {
		panic(err)
	}

	var matches = db.fromSplitRegex.FindAllStringSubmatch(fileContents, -1)

	for _, match := range matches {
		if len(match[2]) > 0 {
			fileContents = strings.Replace(fileContents, match[1], db.fs.ForceTrailingSlash(db.DockerRegistryBasePath)+match[3]+":"+db.Tag, 1)
		}
	}

	// Write out the new file with the tagged base image
	if e := db.fs.WriteFile(
		dynamicDockerfileFilename,
		[]byte(fileContents),
		0644); e != nil {
		panic(e)
	}

	return dynamicDockerfileFilename
}
