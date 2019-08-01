package manifestparser

import (
	"os"
	"path/filepath"
)

type Locator struct {
	FilesToCheckFor []string
}

func NewLocator() *Locator {
	return &Locator{
		FilesToCheckFor: []string{
			"manifest.yml",
			"manifest.yaml",
		},
	}
}

func (loc Locator) GetReadPath(path string) (string, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	resolvedpath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return loc.handleDir(resolvedpath)
	}

	return resolvedpath, nil
}

func (loc Locator) handleDir(dir string) (string, error) {
	for _, filename := range loc.FilesToCheckFor {
		fullPath := filepath.Join(dir, filename)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}

	return "", nil
}
