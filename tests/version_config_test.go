package sdltests

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

type versionConfig struct {
	Disabled *disabledInfo `json:"disabled"`
}

type disabledInfo struct {
	Vulnerabilty bool   `json:"vulnerabilty"`
	Reason       string `json:"reason"`
	Link         string `json:"link"`
}

func TestVersionConfigsAreValid(t *testing.T) {
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

			//subdirectories should be version of that service
			for _, versionDir := range listDirs(t, servicePath) {
				if !versionDir.IsDir() {
					continue
				}

				t.Log("Validating version config for service " + serviceName.Name() + " version " + versionDir.Name())

				configfile := filepath.Join(vendorPath, serviceName.Name(), versionDir.Name(), "config.json")
				contents, fileerr := ioutil.ReadFile(configfile)

				//check if config file exists
				if fileerr == nil {

					assert.Nil(t, fileerr, "Error loading config file for version "+serviceName.Name()+" version "+versionDir.Name())

					//the config json object
					var config versionConfig
					err := json.Unmarshal(contents, &config)
					assert.NoError(t, err, "unmarshalling the version config failed")

					//load the version config and validate it against the schema
					configLoader := gojsonschema.NewReferenceLoader("file://" + configfile)
					configSchemaLoader := gojsonschema.NewReferenceLoader("file://./version_config_schema.json")
					result, err := gojsonschema.Validate(configSchemaLoader, configLoader)
					assert.NoError(t, err, "There was an unexpected error")
					for _, err := range result.Errors() {
						t.Log(err)
					}
					assert.Len(t, result.Errors(), 0, "There was an error validating the version config schema")
				} else {
					t.Log("SDL version does not have a config.json file. Skipping test.")
				}
			}
		}
	}
}
