package cmd

import (
	"github.com/fatih/color"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"encoding/json"

	"gitlab.home.mikenewswanger.com/infrastructure/docker-automatic-build/dockerbuild"
)

// listBaseImagesCmd represents the list command
var listBaseImagesCmd = &cobra.Command{
	Use:   "list-base-images",
	Short: "List Dockerfile Heirarchy",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var db = dockerbuild.DockerBuild{
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
			println(string(output))
		case "yaml":
			var output, err = yaml.Marshal(map[string]interface{}{
				"buildable_images": buildableImages,
				"orphaned_images":  orphanImages,
			})
			if err != nil {
				panic("Failed to marshal json")
			}
			println(string(output))
		default:
			println("Buildable Images:")
			for _, bi := range buildableImages {
				color.Green(bi.Name)
				printImageChildrenStdout(bi, "")
			}
			println("")

			println("Orphaned Images:")
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
