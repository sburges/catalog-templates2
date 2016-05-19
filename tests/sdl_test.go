package sdltests

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
)

type sdl struct {
	Name       string      `json:"name"`
	Version    string      `json:"version"`
	Vendor     string      `json:"vendor"`
	Components []component `json:"components"`
	Parameters []parameter `json:"parameters"`
}

type component struct {
	Name         string               `json:"name"`
	Version      string               `json:"version"`
	Vendor       string               `json:"vendor"`
	Capabilities []string             `json:"capabilities"`
	Parameters   []componentParameter `json:"parameters"`
}

type componentParameter struct {
	Name string `json:"name"`
}

type parameter struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     string `json:"default"`
}

type serviceConfig struct {
	Name                  string   `yaml:"name"`
	Description           string   `yaml:"description"`
	Categories            []string `yaml:"categories"`
	Labels                []string `yaml:"labels"`
	Type                  string   `yaml:"type"`
	DefaultServiceVersion string   `yaml:"default_service_version"`
}

func TestSDLsAreValid(t *testing.T) {
	servicesRoot := "../services"

	for _, vendor := range listDirs(t, servicesRoot) {
		vendorPath := servicesRoot + "/" + vendor.Name()

		//load all services
		for _, serviceName := range listDirs(t, vendorPath) {

			//every directory should be a service
			if !serviceName.IsDir() {
				continue
			}

			servicePath := vendorPath + "/" + serviceName.Name()
			verSet := versionSet(t, servicePath, len(listDirs(t, servicePath)))

			configFilePath := vendorPath + "/" + serviceName.Name() + "/config.yml"
			configfile, configfileerr := ioutil.ReadFile(configFilePath)
			assert.Nil(t, configfileerr, "Error loading config file for service "+serviceName.Name())
			assert.True(t, len(configfile) > 0, "Found empty config file for service "+serviceName.Name())

			//check default version exists
			var s serviceConfig
			err := yaml.Unmarshal(configfile, &s)
			assert.Nil(t, err, "unmarshalling the service config failed")
			if s.DefaultServiceVersion != "" {
				assert.Contains(t, verSet, s.DefaultServiceVersion, "version in service config does not match a dir in tree")
			}
			assert.NotEmpty(t, s.Name, "name must be defined")
			assert.NotEmpty(t, s.Type, "type must be defined")
			assert.NotEmpty(t, s.Description, "description must be defined")
			assert.NotEmpty(t, s.Categories, "categories must be defined")
			assert.NotEmpty(t, s.Labels, "labels must be defined")

			//subdirectories should be version of that service
			for _, versionDir := range listDirs(t, servicePath) {
				if !versionDir.IsDir() {
					continue
				}

				t.Log("Validating Service " + serviceName.Name() + " version " + versionDir.Name())

				sdlfile := vendorPath + "/" + serviceName.Name() + "/" + versionDir.Name() + "/sdl/sdl.json"
				contents, fileerr := ioutil.ReadFile(sdlfile)

				//sdl should exist
				assert.Nil(t, fileerr, "Error loading sdl file for service "+serviceName.Name()+" version "+versionDir.Name())

				//sdl should not be empty
				assert.True(t, len(contents) > 0, "Found empty sdl file for service "+serviceName.Name()+" version "+versionDir.Name())

				//the sdl json object
				var m sdl
				err := json.Unmarshal(contents, &m)
				assert.Nil(t, err, "unmarshalling the sdl failed")

				//check contents match version and vendor directories
				assert.Equal(t, vendor.Name(), m.Vendor, "vendor name does not match in sdl and directory")
				assert.Contains(t, verSet, m.Version, "version in sdl does not match dir its in")

				//convert sdl parameters to componentParameter type for check
				var sdlParams []componentParameter
				for _, param := range m.Parameters {
					sdlParams = append(sdlParams, componentParameter{Name: param.Name})

					//check that there is no default value set for parameters that contain 'PASSWORD' in the name
					if strings.Contains(strings.ToLower(param.Name), "password") {
						assert.Empty(t, param.Default, "SDL parameters that are passwords should not have a default value set")
					}
				}

				var compParams []componentParameter
				var compNames []string
				for _, comp := range m.Components {
					compNames = append(compNames, comp.Name)

					//log a warning if component has "ALL" capabilities set
					for _, cap := range comp.Capabilities {
						if cap == "ALL" {
							t.Log("***** WARNING - " + comp.Name + " - should verify that ALL capabilities is required")
						}
					}

					//check component parameters exist in parameter definition list
					for _, param := range comp.Parameters {
						assert.Contains(t, sdlParams, param, "component parameter not found in sdl parameter definition list")
						compParams = append(compParams, param)
					}
				}

				//check that a csm-side-car container exists in components
				assert.Contains(t, compNames, "csm-side-car", "SDL needs a csm-side-car defined")

				//check the parameter name and description are not equal and that all SDL parameters are used in components
				for _, param := range m.Parameters {
					assert.NotEqual(t, param.Name, param.Description, "Parameter name and description should not be the same")
					assert.Contains(t, compParams, componentParameter{Name: param.Name}, "SDL parameter defined but not used in a component")
				}

				//load the sdl and validate it against the schema
				sdlLoader := gojsonschema.NewReferenceLoader("file://" + sdlfile)
				sdlSchemaLoader := gojsonschema.NewReferenceLoader("file://./sdl_schema.json")
				result, err := gojsonschema.Validate(sdlSchemaLoader, sdlLoader)
				assert.Nil(t, err, "There was an unexpected error")

				for _, err := range result.Errors() {
					t.Log(err)
				}

				assert.Equal(t, len(result.Errors()), 0, "There was an error validating the SDL schema")
			}
		}
	}
}

func listDirs(t *testing.T, path string) []os.FileInfo {
	dirList, err := ioutil.ReadDir(path)
	assert.Nil(t, err, "Error listing services directory")
	return dirList
}

func versionSet(t *testing.T, versionDir string, size int) map[string]struct{} {
	vs := make(map[string]struct{}, size)
	for _, d := range listDirs(t, versionDir) {
		vs[d.Name()] = struct{}{}
	}
	return vs
}
