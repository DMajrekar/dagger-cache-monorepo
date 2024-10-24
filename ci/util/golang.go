package util

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"

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

	return dag.Container().
		From(baseContainerVersion).
		WithMountedCache("/go/pkg/mod", cacheVolume, dagger.ContainerWithMountedCacheOpts{Sharing: dagger.Shared}).
		WithMountedCache("/root/.cache/go-build", cacheBuild).
		WithDirectory("/go/src/github.com/dmajrekar/dagger-cache-monorepo/"+dir, rootDir.Directory(dir)).
		WithDirectory("/go/src/github.com/dmajrekar/dagger-cache-monorepo/lib", rootDir.Directory("lib")).
		WithWorkdir("/go/src/github.com/dmajrekar/dagger-cache-monorepo/" + dir).
		WithExec([]string{"go", "run", "main.go"}).
		Stdout(ctx)
}
