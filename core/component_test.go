package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelativePathToGitComponent(t *testing.T) {
	subcomponent := Component{
		Name:   "efk",
		Method: "git",
		Source: "https://github.com/Microsoft/fabrikate-elasticsearch-fluentd-kibana",
	}

	assert.Equal(t, subcomponent.RelativePathTo(), "components/efk")
}

func TestRelativePathToDirectoryComponent(t *testing.T) {
	subcomponent := Component{
		Name:   "infra",
		Source: "./infra",
	}

	assert.Equal(t, subcomponent.RelativePathTo(), "./infra")
}

func TestLoadComponent(t *testing.T) {
	component := Component{
		PhysicalPath: "../test/fixtures/definition/infra",
		LogicalPath:  "infra",
	}

	component, err := component.LoadComponent()
	assert.Nil(t, err)

	assert.Nil(t, err)
	assert.Equal(t, component.Name, "infra")
	assert.Equal(t, len(component.Subcomponents), 1)
	assert.Equal(t, component.Subcomponents[0].Name, "efk")
	assert.Equal(t, component.Subcomponents[0].Source, "https://github.com/Microsoft/fabrikate-elasticsearch-fluentd-kibana")
	assert.Equal(t, component.Subcomponents[0].Method, "git")
}

func TestLoadConfig(t *testing.T) {
	component := Component{
		PhysicalPath: "../test/fixtures/generate/infra",
		LogicalPath:  "infra",
	}

	component, err := component.LoadComponent()
	assert.Nil(t, err)

	err = component.LoadConfig([]string{"prod-east", "prod"})

	assert.Nil(t, err)
}

func TestIteratingDefinition(t *testing.T) {
	callbackCount := 0
	results := WalkComponentTree("../test/fixtures/iterator", []string{""}, func(path string, component *Component) (err error) {
		callbackCount++
		return nil
	})

	var err error
	components := make([]Component, 0)
	for result := range results {
		if result.Error != nil {
			err = result.Error
		} else if result.Component != nil {
			components = append(components, *result.Component)
		}
	}

	assert.Nil(t, err)
	assert.Equal(t, 3, len(components))
	assert.Equal(t, callbackCount, len(components))

	assert.Equal(t, components[1].PhysicalPath, "../test/fixtures/iterator/infra")
	assert.Equal(t, components[1].LogicalPath, "infra")

	assert.Equal(t, components[2].PhysicalPath, "../test/fixtures/iterator/infra/components/efk")
	assert.Equal(t, components[2].LogicalPath, "infra/efk")
}

func TestWriteComponent(t *testing.T) {
	component := Component{
		PhysicalPath: "../test/fixtures/install",
		LogicalPath:  "",
	}

	component, err := component.LoadComponent()
	assert.Nil(t, err)

	err = component.Write()
	assert.Nil(t, err)
}
