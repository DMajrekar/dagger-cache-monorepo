package util

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"
	"strings"

	"golang.org/x/mod/modfile"
)

func Run(ctx context.Context, dag *dagger.Client, dir string, rootDir *dagger.Directory) (string, error) {
	cacheVolume := dag.CacheVolume("go-mod-cache-" + dir)
	cacheBuild := dag.CacheVolume("go-build-cache-" + dir)

	projectDir := rootDir.Directory(dir)

	// fetch go mod file
	goModFile, err := projectDir.File("go.mod").Contents(ctx)
	if err != nil {
		return "", err
	}

	// Parse the go.mod file
	mod, err := modfile.Parse("", []byte(goModFile), nil)
	if err != nil {
		return "", err
	}

	// Set the base container version from the go version in the go.mod file
	baseContainerVersion := fmt.Sprintf("golang:%s", mod.Go.Version)
	baseContainer := dag.Container().From(baseContainerVersion)

	replacedMods, err := replacedDepsFromGoMod(goModFile)
	if err != nil {
		return "", err
	}
	for _, r := range replacedMods {
		// Mount the replaced dependencies
		mountPath := strings.Replace(r, "../", "", 1)
		baseContainer = baseContainer.WithDirectory("/go/src/github.com/dmajrekar/dagger-cache-monorepo/"+mountPath, rootDir.Directory(mountPath))
	}

	baseContainer = baseContainer.
		WithMountedCache("/go/pkg/mod", cacheVolume, dagger.ContainerWithMountedCacheOpts{Sharing: dagger.Shared}).
		WithMountedCache("/root/.cache/go-build", cacheBuild).
		WithEnvVariable("GOPRIVATE", "git/*").
		WithDirectory("/go/src/github.com/dmajrekar/dagger-cache-monorepo/"+dir, rootDir.Directory(dir)).
		WithWorkdir("/go/src/github.com/dmajrekar/dagger-cache-monorepo/" + dir)

	return baseContainer.
		WithExec([]string{"/bin/sh", "-c", "go run main.go"}).
		Stdout(ctx)
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
