package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/manifestparser"
	"github.com/cloudfoundry/bosh-cli/director/template"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

//go:generate counterfeiter . ManifestParser

type ManifestParser interface {
	Apps() []manifestparser.Application
	ContainsManifest() bool
	FullRawManifest() []byte
}

type Parser struct {
	Manifest       ManifestData
	pathToManifest string
	rawManifest    []byte
	hasParsed      bool
}

type ParsedManifest interface {
	HasMultipleApps() bool
	ForApp(appName string) (ParsedManifest, error)
	OverrideFirstAppName(appName string) ParsedManifest
	OverrideFirstAppBuildpacks(buildpacks []string) ParsedManifest

	OverrideFirstAppStack(stack string) ParsedManifest
	//OverrideDisk(value types.NullUint64)
	//OverrideDropletPath(value string)
	//OverrideDockerImage(value string)
	//OverrideDockerPassword(value string)
	//OverrideDockerUsername(value string)
	//OverrideHealthCheckEndpoint(value string)
	//OverrideHealthCheckTimeout(value int64)
	//OverrideHealthCheckType(value constant.HealthCheckType)
	//OverrideInstances(value types.NullInt)
	//OverrideMemory(value types.NullUint64)
	//OverrideNoRoute(value bool)
	//OverrideRandomRoute(value bool)
	//OverrideStartCommand(value types.FilteredString)
}

type ManifestData struct {
	Applications []manifestparser.Application `yaml:"applications"`
}

func (parser *Parser) InterpolateAndParse(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) error {
	rawManifest, err := ioutil.ReadFile(pathToManifest)
	if err != nil {
		return err
	}

	tpl := template.NewTemplate(rawManifest)
	fileVars := template.StaticVariables{}

	for _, path := range pathsToVarsFiles {
		rawVarsFile, ioerr := ioutil.ReadFile(path)
		if ioerr != nil {
			return ioerr
		}

		var sv template.StaticVariables

		err = yaml.Unmarshal(rawVarsFile, &sv)
		if err != nil {
			return manifestparser.InvalidYAMLError{Err: err}
		}

		for k, v := range sv {
			fileVars[k] = v
		}
	}

	for _, kv := range vars {
		fileVars[kv.Name] = kv.Value
	}

	rawManifest, err = tpl.Evaluate(fileVars, nil, template.EvaluateOpts{ExpectAllKeys: true})
	if err != nil {
		return manifestparser.InterpolationError{Err: err}
	}

	parser.pathToManifest = pathToManifest
	return parser.parse(rawManifest)
}

func (parser *Parser) parse(manifestBytes []byte) error {
	parser.rawManifest = manifestBytes

	var raw ManifestData
	err := yaml.Unmarshal(manifestBytes, &raw)
	if err != nil {
		return err
	}

	parser.Manifest = raw
	parser.hasParsed = true
	return nil
}

func (parser *Parser) GetManifest(noManifest bool, pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) (ParsedManifest, error) {
	if noManifest || pathToManifest == "" {
		return ManifestData{Applications: []manifestparser.Application{{}}}, nil
	}

	err := parser.InterpolateAndParse(pathToManifest, pathsToVarsFiles, vars)

	return parser.Manifest, err
}

func (manifest ManifestData) HasMultipleApps() bool {
	return len(manifest.Applications) > 1
}

func (manifest ManifestData) ForApp(appName string) (ParsedManifest, error) {
	for _, appSection := range manifest.Applications {
		if appSection.Name == appName {
			return ManifestData{Applications: []manifestparser.Application{appSection}}, nil
		}
	}

	return ManifestData{}, actionerror.AppNotFoundInManifestError{Name: appName}
}

func (manifest ManifestData) OverrideFirstAppName(appName string) ParsedManifest {
	manifest.Applications[0].Name = appName
	return manifest
}

func (manifest ManifestData) OverrideFirstAppBuildpacks(buildpacks []string) ParsedManifest {
	manifest.Applications[0].Buildpacks = buildpacks
	return manifest
}

func (manifest ManifestData) OverrideFirstAppStack(stack string) ParsedManifest {
	manifest.Applications[0].Stack = stack
	return manifest
}

func (manifest ManifestData) OverrideFirstAppDisk(disk types.NullUint64) ParsedManifest {
	if disk.IsSet {
		*(manifest.Applications[0].DiskQuota) = uint(disk.Value)
	}

	return manifest
}
