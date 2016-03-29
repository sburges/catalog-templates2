package sdltests

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/xeipuuv/gojsonschema"
    "io/ioutil"
)

func TestSDLsAreValid(t *testing.T) {
    servicepath := "../services"
    dirList, err := ioutil.ReadDir(servicepath)
    assert.Nil(t, err, "Error listing services directory")
    
    //load all services 
    for _, file := range dirList {
        
        //every directory should be a service
        if file.IsDir() {
            t.Log("Validating Service: " + file.Name())
            innerList, innererr := ioutil.ReadDir(servicepath + "/" + file.Name())
            assert.Nil(t, innererr, "Error listing version of service: " + file.Name())
            
            configfile, configfileerr := ioutil.ReadFile(servicepath + "/" + file.Name() + "/config.yml")
            assert.Nil(t, configfileerr, "Error loading config file for service " + file.Name())
            assert.True(t, len(configfile) > 0, "Found empty config file for service " + file.Name())
                                
            //subdirectories should be version of that service
            for _, subFile := range innerList {
                if subFile.IsDir() {
                    t.Log("Validating Service " + file.Name() + " version " + subFile.Name())
                    
                    sdlfile := servicepath + "/" + file.Name() + "/" + subFile.Name() + "/sdl/sdl.json"
                    contents, fileerr := ioutil.ReadFile(sdlfile)
                    
                    //sdl should exist
                    assert.Nil(t, fileerr, "Error loading sdl file for service " + file.Name() + " version " + subFile.Name())
                    
                    //sdl should not be empty
                    assert.True(t, len(contents) > 0, "Found empty sdl file for service " + file.Name() + " version " + subFile.Name())
                    
                    //load the sdl and validate it against the schema
                    documentLoader := gojsonschema.NewReferenceLoader("file://" + sdlfile)
                    schemaLoader := gojsonschema.NewReferenceLoader("file://./sdl_schema.json")
                    
                    result,err := gojsonschema.Validate(schemaLoader, documentLoader)
                    assert.Nil(t, err, "There was an unexpected error")
                    
                    for _, err := range result.Errors() {
                        t.Log(err)
                    }
        
                    assert.Equal(t, len(result.Errors()), 0, "There was an error validating the SDL schema") 
                }
                
            }
        }
    }
}