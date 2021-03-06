package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/Microsoft/fabrikate/core"
	"github.com/Microsoft/fabrikate/generators"
	"github.com/kyokomi/emoji"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func WriteGeneratedManifests(generationPath string, components []core.Component) (err error) {
	// Delete the old version, so we don't end up with a mishmash of two builds.
	os.RemoveAll(generationPath)

	for _, component := range components {
		componentGenerationPath := path.Join(generationPath, component.LogicalPath)
		err := os.MkdirAll(componentGenerationPath, 0755)
		if err != nil {
			return err
		}

		componentYAMLFilename := fmt.Sprintf("%s.yaml", component.Name)
		componentYAMLFilePath := path.Join(componentGenerationPath, componentYAMLFilename)

		log.Info(emoji.Sprintf(":floppy_disk: Writing %s", componentYAMLFilePath))

		err = ioutil.WriteFile(componentYAMLFilePath, []byte(component.Manifest), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func ValidateGeneratedManifests(generationPath string) (err error) {
	log.Println(emoji.Sprintf(":microscope: validating generated manifests in path %s", generationPath))
	output, err := exec.Command("kubectl", "apply", "--validate=true", "--dry-run", "--recursive", "-f", generationPath).Output()

	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			log.Errorf("validating generated manifests failed with: %s: output: %s\n", ee.Stderr, output)
			return err
		}
	}

	return nil
}

func Generate(startPath string, environments []string, validate bool) (components []core.Component, err error) {
	// Iterate through component tree and generate
	results := core.WalkComponentTree(startPath, environments, func(path string, component *core.Component) (err error) {

		var generator core.Generator
		switch component.Generator {
		case "helm":
			generator = &generators.HelmGenerator{}
		case "static":
			generator = &generators.StaticGenerator{}
		}

		return component.Generate(generator)
	})

	components, err = core.SynchronizeWalkResult(results)

	if err != nil {
		return nil, err
	}

	environmentName := strings.Join(environments, "-")
	if len(environmentName) == 0 {
		environmentName = "common"
	}

	generationPath := path.Join(startPath, "generated", environmentName)

	if err = WriteGeneratedManifests(generationPath, components); err != nil {
		return nil, err
	}

	if validate {
		if err = ValidateGeneratedManifests(generationPath); err != nil {
			return nil, err
		}
	}

	if err == nil {
		log.Info(emoji.Sprintf(":raised_hands: finished generate"))
	}

	return components, err
}

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate <config1> <config2> ... <configN>",
	Short: "Generates Kubernetes resource definitions from deployment definition.",
	Long: `Generate produces Kubernetes resource manifests from a deployment definition.

If multiple configurations are specified, each will be applied in left to right priority order at each level of the definition, and the final generated environment directory will have the form <config1>-<config2>-...-<configN>.

For example, 'fab generate prod east' will generate to a directory named prod-east and 'prod' will take priority over 'east'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		PrintVersion()

		validation := cmd.Flag("validate").Value.String()
		_, err := Generate("./", args, validation == "true")

		return err
	},
}

func init() {
	generateCmd.PersistentFlags().Bool("validate", false, "Validate generated resource manifest YAML")
	rootCmd.AddCommand(generateCmd)
}
