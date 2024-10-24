package util

import (
	"context"
	"dagger/ci/internal/dagger"
)

func Run(ctx context.Context, dag *dagger.Client, dir string, rootDir *dagger.Directory) (string, error) {
	cacheVolume := dag.CacheVolume("go-mod-cache-" + dir)
	cacheBuild := dag.CacheVolume("go-build-cache-" + dir)

	return dag.Container().
		From("golang:1.23.2").
		WithMountedCache("/go/pkg/mod", cacheVolume, dagger.ContainerWithMountedCacheOpts{Sharing: dagger.Shared}).
		WithMountedCache("/root/.cache/go-build", cacheBuild).
		WithDirectory("/go/src/github.com/dmajrekar/dagger-cache-monorepo/"+dir, rootDir.Directory(dir)).
		WithDirectory("/go/src/github.com/dmajrekar/dagger-cache-monorepo/lib", rootDir.Directory("lib")).
		WithWorkdir("/go/src/github.com/dmajrekar/dagger-cache-monorepo/" + dir).
		WithExec([]string{"go", "run", "main.go"}).
		Stdout(ctx)
}
