package manifestparser

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/types"
)

type Manifest struct {
	Applications []Application `yaml:"applications"`
}

func (manifest Manifest) HasMultipleApps() bool {
	return len(manifest.Applications) > 1
}

func (manifest Manifest) ForApp(appName string) (Manifest, error) {
	for _, appSection := range manifest.Applications {
		if appSection.Name == appName {
			return Manifest{Applications: []Application{appSection}}, nil
		}
	}

	return Manifest{}, actionerror.AppNotFoundInManifestError{Name: appName}
}

func (manifest Manifest) OverrideFirstAppName(appName string) Manifest {
	manifest.Applications[0].Name = appName
	return manifest
}

func (manifest Manifest) OverrideFirstAppBuildpacks(buildpacks []string) Manifest {
	manifest.Applications[0].Buildpacks = buildpacks
	return manifest
}

func (manifest Manifest) OverrideFirstAppStack(stack string) Manifest {
	manifest.Applications[0].Stack = stack
	return manifest
}

func (manifest Manifest) OverrideFirstAppDisk(disk types.NullUint64) Manifest {
	if disk.IsSet {
		*(manifest.Applications[0].DiskQuota) = uint(disk.Value)
	}

	return manifest
}
