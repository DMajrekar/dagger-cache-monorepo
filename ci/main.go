package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"
	"strings"
)

type Ci struct{}

func (Ci) RunTest(ctx context.Context,
	// +defaultPath="/"
	// +ignore=["*", "!project*", "!lib*" ]
	rootDir *dagger.Directory,
) (string, error) {
	dirs, err := filterDirs(ctx, rootDir)
	if err != nil {
		return "", err
	}

	for _, dir := range dirs {
		fmt.Printf("Running test for %s\n", dir)

		cacheVolume := dag.CacheVolume("go-mod-cache-" + dir)
		cacheBuild := dag.CacheVolume("go-build-cache-" + dir)

		out, err := dag.Container().
			From("golang:1.23.2").
			WithMountedCache("/go/pkg/mod", cacheVolume, dagger.ContainerWithMountedCacheOpts{Sharing: dagger.Shared}).
			WithMountedCache("/root/.cache/go-build", cacheBuild).
			WithDirectory("/go/src/github.com/dmajrekar/dagger-cache-monorepo/"+dir, rootDir.Directory(dir)).
			WithDirectory("/go/src/github.com/dmajrekar/dagger-cache-monorepo/lib", rootDir.Directory("lib")).
			WithWorkdir("/go/src/github.com/dmajrekar/dagger-cache-monorepo/" + dir).
			WithExec([]string{"go", "run", "main.go"}).
			Stdout(ctx)
		if err != nil {
			return "", err
		}

		fmt.Printf("Output \n%s\n", out)
	}

	return "", nil
}

// filterDirs will return a list of directories within the srcDir
func filterDirs(ctx context.Context, srcDir *dagger.Directory) ([]string, error) {
	filterDirsList := []string{}

	files, err := srcDir.Entries(ctx)
	if err != nil {
		return filterDirsList, err
	}

	for _, file := range files {

		_, err := srcDir.Directory(file).Entries(ctx)
		if err != nil {
			continue
		}

		// if it doesn't start with project, skip
		if !strings.HasPrefix(file, "project") {
			continue
		}

		filterDirsList = append(filterDirsList, file)
	}

	return filterDirsList, nil
}
