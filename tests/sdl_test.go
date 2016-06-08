package sdltests

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
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
	DeploymentGrade       string   `yaml:"deployment_grade"`
	Supported             string   `yaml:"supported"`
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

			servicePath := filepath.Join(vendorPath, serviceName.Name())
			verSet := versionSet(t, servicePath, len(listDirs(t, servicePath)))

			configFilePath := filepath.Join(vendorPath, serviceName.Name(), "config.yml")
			configfile, configfileerr := ioutil.ReadFile(configFilePath)
			assert.Nil(t, configfileerr, "Error loading config file for service "+serviceName.Name())
			assert.True(t, len(configfile) > 0, "Found empty config file for service "+serviceName.Name())

			//check default version exists
			var c serviceConfig
			err := yaml.Unmarshal(configfile, &c)
			assert.Nil(t, err, "unmarshalling the service config failed")
			assert.NotEmpty(t, c.Name, "name must be defined")
			assert.NotEmpty(t, c.Type, "type must be defined")
			assert.NotEmpty(t, c.Description, "description must be defined")
			assert.NotEmpty(t, c.Categories, "categories must be defined")
			assert.NotEmpty(t, c.Labels, "labels must be defined")
			if len(verSet) > 1 {
				assert.NotEmpty(t, c.DefaultServiceVersion, "found more than 1 version and default_service_version is not set in config.yml")
			}
			if c.DefaultServiceVersion != "" {
				assert.Contains(t, verSet, c.DefaultServiceVersion, "version in service config does not match a dir in tree")
			}
			// DeploymentGrade should either be production or development or nil
			if c.DeploymentGrade != "" {
				deployGrade := strings.ToLower(c.DeploymentGrade)
				assert.True(t, deployGrade == "production" || deployGrade == "development", "deployment grade can be empty or development or production")
			}
			// Supported should either be production or development or nil
			if c.Supported != "" {
				support := strings.ToLower(c.Supported)
				assert.True(t, support == "true" || support == "false", "support can be empty or true or false")
			}

			//subdirectories should be version of that service
			for _, versionDir := range listDirs(t, servicePath) {
				if !versionDir.IsDir() {
					continue
				}

				t.Log("Validating Service " + serviceName.Name() + " version " + versionDir.Name())

				sdlfile := filepath.Join(vendorPath, serviceName.Name(), versionDir.Name(), "sdl/sdl.json")
				contents, fileerr := ioutil.ReadFile(sdlfile)

				//sdl should exist
				assert.Nil(t, fileerr, "Error loading sdl file for service "+serviceName.Name()+" version "+versionDir.Name())

				//sdl should not be empty
				assert.True(t, len(contents) > 0, "Found empty sdl file for service "+serviceName.Name()+" version "+versionDir.Name())

				//the sdl json object
				var s sdl
				err := json.Unmarshal(contents, &s)
				assert.Nil(t, err, "unmarshalling the sdl failed")

				//check contents match version and vendor directories
				assert.Equal(t, vendor.Name(), s.Vendor, "vendor name does not match in sdl and directory")
				assert.Contains(t, verSet, s.Version, "version in sdl does not match dir its in")

				//convert sdl parameters to componentParameter type for check
				var sdlParams []componentParameter
				for _, param := range s.Parameters {
					sdlParams = append(sdlParams, componentParameter{Name: param.Name})

					//check that there is no default value set for parameters that contain 'PASSWORD' in the name
					if strings.Contains(strings.ToLower(param.Name), "password") {
						assert.Empty(t, param.Default, "SDL parameters that are passwords should not have a default value set")
					}
				}

				var compParams []componentParameter
				var compNames []string
				for _, comp := range s.Components {
					compNames = append(compNames, comp.Name)

					if comp.Name == "csm-side-car" {
						p := componentParameter{Name: "CSM_API_KEY"}
						assert.Contains(t, comp.Parameters, p, "CSM_API_KEY is not present in csm-side-car component")
					}

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
				for _, param := range s.Parameters {
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
		if !d.IsDir() {
			continue
		}
		vs[d.Name()] = struct{}{}
	}
	return vs
}
