package manifestparser

import (
	"reflect"
	"strings"
)

// Any manifest fields which are affected by flag overrides are fields on this struct
// any unkown fields will be stored in the map.
type Application struct {
	Name                    string   `yaml:"name"`
	Buildpacks              []string `yaml:"buildpacks"`
	DiskQuota               *uint    `yaml:"disk_quota"`
	Docker                  *Docker  `yaml:"docker"`
	Path                    string   `yaml:"path"`
	NoRoute                 bool     `yaml:"no-route"`
	RandomRoute             bool     `yaml:"random-route"`
	Stack                   string   `yaml:"stack"`
	RemainingManifestFields map[string]interface{}
}

func (application *Application) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	err := unmarshal(&application.RemainingManifestFields)
	if err != nil {
		return err
	}
	err = unmarshal(&application)
	if err != nil {
		return err
	}

	// Remove struct fields from map so we don't duplicate manifest fields
	value := reflect.ValueOf(application)
	for i := 0; i < value.NumField(); i++ {
		structField := value.Type().Field(i)

		yamlTag := strings.Split(structField.Tag.Get("yaml"), ",")

		yamlKey := structField.Name
		if nameFromTag := yamlTag[0]; nameFromTag != "" {
			yamlKey = nameFromTag
		}

		delete(application.RemainingManifestFields, yamlKey)
	}
	return nil
}

type Docker struct {
	Image    string `yaml:"image"`
	Username string `yaml:"username"`
}
