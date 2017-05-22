package cmd

import (
    "encoding/json"

	"github.com/fatih/color"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"go.mikenewswanger.com/docker-automatic-build/dockerbuild"
)

// listBaseImagesCmd represents the list command
var listBaseImagesCmd = &cobra.Command{
	Use:   "list-base-images",
	Short: "List Dockerfile Heirarchy",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var db = dockerbuild.DockerBuild{
			Logger:                 logger,
			Verbosity:              uint8(commandLineFlags.verbosity),
			DockerBaseDirectory:    commandLineFlags.dockerBaseDirectory,
			DockerRegistryBasePath: commandLineFlags.dockerRegistryBasePath,
		}

		var buildableImages, orphanImages = db.GetBaseImageHeirarchy()

		switch commandLineFlags.outputFormat {
		case "json":
			var output, err = json.Marshal(map[string]interface{}{
				"buildable_images": buildableImages,
				"orphaned_images":  orphanImages,
			})
			if err != nil {
				panic("Failed to marshal json")
			}
			color.White(string(output))
		case "yaml":
			var output, err = yaml.Marshal(map[string]interface{}{
				"buildable_images": buildableImages,
				"orphaned_images":  orphanImages,
			})
			if err != nil {
				panic("Failed to marshal json")
			}
			color.White(string(output))
		default:
			color.White("Buildable Images:")
			for _, bi := range buildableImages {
				color.Green(bi.Name)
				printImageChildrenStdout(bi, "")
			}
			color.White("")

			color.White("Orphaned Images:")
			for _, oi := range orphanImages {
				color.Red(oi.Name + " (Missing parent: " + oi.ParentName + ")")
			}
		}
	},
}

func printImageChildrenStdout(dbi dockerbuild.DockerBuildableImage, prefix string) {
	for _, c := range dbi.Children {
		color.Green(prefix + "  ↳" + c.Name)
		printImageChildrenStdout(c, "   "+prefix)
	}
}

func init() {
	RootCmd.AddCommand(listBaseImagesCmd)

	listBaseImagesCmd.Flags().StringVarP(&commandLineFlags.outputFormat, "output-format", "o", "", "Specify output format.  Available options are stdout (default), json, and yaml")
}
