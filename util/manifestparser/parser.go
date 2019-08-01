package manifestparser

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-cli/director/template"
	"gopkg.in/yaml.v2"
)

// TODO I dont think this needs state if we just treat it as a way
// to change manifest struct <-> bytes
type Parser struct{}

func NewParser() *Parser {
	return new(Parser)
}

func (parser *Parser) GetManifest(noManifest bool, pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) (Manifest, error) {
	if noManifest || pathToManifest == "" {
		return Manifest{Applications: []Application{{}}}, nil
	}

	return parser.InterpolateAndParse(pathToManifest, pathsToVarsFiles, vars)
}

// This should marshal the manifest and return the bytes
func (parser *Parser) MarshallManifest(manifest Manifest) ([]byte, error) {
	return yaml.Marshal(manifest)
}

// TODO: make this private push should enter through get manifest
// and apply manifest should error if there is not path so it shouldnt reach this point

// InterpolateAndParse reads the manifest at the provided paths, interpolates
// variables if a vars file is provided, and sets the current manifest to the
// resulting manifest.
// For manifests with only 1 application, appName will override the name of the
// single app defined.
// For manifests with multiple applications, appName will filter the
// applications and leave only a single application in the resulting parsed
// manifest structure.
func (parser *Parser) InterpolateAndParse(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) (Manifest, error) {
	rawManifest, err := ioutil.ReadFile(pathToManifest)
	if err != nil {
		return Manifest{}, err
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
			return Manifest{}, InvalidYAMLError{Err: err}
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
		return Manifest{}, InterpolationError{Err: err}
	}

	return parser.parse(rawManifest, pathToManifest)
}

// TODO I removed name presence validation in this method we need to validate somewhere else
// It felt weird for the parser to know about an appName. It should just read in a manifest
// and push will do its own validations

// TODO comapre to master and make sure we didnt lost any validation coverage from
// trimming this down
func (parser *Parser) parse(manifestBytes []byte, pathToManifest string) (Manifest, error) {
	var manifestStruct Manifest

	err := yaml.Unmarshal(manifestBytes, &manifestStruct)
	if err != nil {
		return Manifest{}, err
	}

	if len(manifestStruct.Applications) == 0 {
		return Manifest{}, errors.New("must have at least one application")
	}

	for i := range manifestStruct.Applications {

		if manifestStruct.Applications[i].Path == "" {
			continue
		}

		var finalPath = manifestStruct.Applications[i].Path
		if !filepath.IsAbs(finalPath) {
			finalPath = filepath.Join(filepath.Dir(pathToManifest), finalPath)
		}
		finalPath, err = filepath.EvalSymlinks(finalPath)
		if err != nil {
			if os.IsNotExist(err) {
				return Manifest{}, InvalidManifestApplicationPathError{
					Path: manifestStruct.Applications[i].Path,
				}
			}
			return Manifest{}, err
		}
		manifestStruct.Applications[i].Path = finalPath
	}

	return manifestStruct, nil
}
