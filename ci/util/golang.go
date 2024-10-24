package util

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"
	"strings"

	"golang.org/x/mod/modfile"
)

func RunGoTest(ctx context.Context, client *dagger.Client, projectBaseDir string, neoRootDir *dagger.Directory) (*dagger.Container, error) {
	// Fetch the project directory
	projectDir := neoRootDir.Directory(projectBaseDir)

	// fetch go mod file
	goModFile, err := projectDir.File("go.mod").Contents(ctx)
	if err != nil {
		return nil, err
	}

	// Parse the go.mod file
	mod, err := modfile.Parse("", []byte(goModFile), nil)
	if err != nil {
		return nil, err
	}

	// Set the base container version from the go version in the go.mod file
	baseContainerVersion := fmt.Sprintf("golang:%s", mod.Go.Version)
	baseContainer := client.Container().From(baseContainerVersion)

	// Fetch a list of all replaced dependencies in the go.mod file
	replacedDeps, err := replacedDepsFromGoMod(goModFile)
	if err != nil {
		return nil, err
	}

	// For each replaced dependency, add the directory to the container
	for _, dep := range replacedDeps {
		// replace ../ from dep
		dep = strings.Replace(dep, "../", "", 1)
		baseContainer = baseContainer.WithDirectory("/go/src/github.com/dmajrekar/dagger-cache-monorepo/"+dep, neoRootDir.Directory(dep))
	}

	// create a dedicated cache volume for the project
	cacheVolume := client.CacheVolume(projectBaseDir)

	// Add the project directory to the container, with caches/ env vars
	baseContainer = baseContainer.
		WithMountedCache("/go/pkg/mod", cacheVolume, dagger.ContainerWithMountedCacheOpts{Sharing: dagger.Shared}).
		WithMountedCache("/root/.cache/go-build", client.CacheVolume("go-build-cache")).
		WithEnvVariable("GOPRIVATE", "git/*").
		WithDirectory("/go/src/github.com/dmajrekar/dagger-cache-monorepo/"+projectBaseDir, projectDir).
		WithWorkdir("/go/src/github.com/dmajrekar/dagger-cache-monorepo/" + projectBaseDir)

	// Run the main.go
	return baseContainer.
		WithExec([]string{"/bin/sh", "-c", "go run main.go"},
			dagger.ContainerWithExecOpts{
				InsecureRootCapabilities:      true,
				ExperimentalPrivilegedNesting: true,
			}).
		Sync(ctx)
}

// replacedDepsFromGoMod takes a go.mod file and returns the list of
// replaced dependencies
func replacedDepsFromGoMod(goModFile string) ([]string, error) {
	mod, err := modfile.Parse("", []byte(goModFile), nil)
	if err != nil {
		return nil, err
	}

	var deps []string
	for _, r := range mod.Replace {
		// exit if the replace is not a local path
		if r.New.Version != "" {
			continue
		}
		deps = append(deps, r.New.Path)
	}

	return deps, nil
}
