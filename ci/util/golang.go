package util

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"
	"strings"

	"golang.org/x/mod/modfile"
)

func RunGoTest(ctx context.Context, client *dagger.Client, projectBaseDir string, neoRootDir *dagger.Directory) (*dagger.Container, error) {
	projectDir := neoRootDir.Directory(projectBaseDir)
	testContainer, err := baseGoContainer(ctx, client, projectBaseDir, neoRootDir, projectDir)
	if err != nil {
		return nil, err
	}

	testContainer = testContainer.
		WithExec([]string{"/bin/sh", "-c", "go run main.go"},
			dagger.ContainerWithExecOpts{
				InsecureRootCapabilities:      true,
				ExperimentalPrivilegedNesting: true,
			})
	_, err = testContainer.Stdout(ctx)
	// Ensure that the tests have been run
	return testContainer, err
}

func baseGoContainer(ctx context.Context, client *dagger.Client, projectPath string, neoRootDir, projectDir *dagger.Directory) (*dagger.Container, error) {
	// fetch go mod file
	goModFile, err := projectDir.File("go.mod").Contents(ctx)
	if err != nil {
		return nil, err
	}

	// base go container of the correct version
	baseContainerVersion, err := GoContainerFromGoMod(goModFile)
	if err != nil {
		return nil, err
	}

	baseContainer := client.Container().From(baseContainerVersion)

	cacheVolume := client.CacheVolume(projectPath)

	testContainer, err := replaceNeoGoModules(goModFile, baseContainer, neoRootDir)
	if err != nil {
		return nil, err
	}
	goContainer := goContainerOpts(client, testContainer, cacheVolume, projectPath, projectDir)

	return goContainer, err
}

func goContainerOpts(client *dagger.Client, baseContainer *dagger.Container, cacheVolume *dagger.CacheVolume, projectPath string, projectDir *dagger.Directory) *dagger.Container {
	return baseContainer.
		WithMountedCache("/go/pkg/mod", cacheVolume, dagger.ContainerWithMountedCacheOpts{Sharing: dagger.Shared}).
		WithMountedCache("/root/.cache/go-build", client.CacheVolume("go-build-cache")).
		WithEnvVariable("GOPRIVATE", "git/*").
		WithDirectory("/go/src/github.com/dmajrekar/dagger-cache-monorepo/"+projectPath, projectDir).
		WithWorkdir("/go/src/github.com/dmajrekar/dagger-cache-monorepo/" + projectPath)

}

func GoContainerFromGoMod(goModFile string) (string, error) {
	mod, err := modfile.Parse("", []byte(goModFile), nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("golang:%s", mod.Go.Version), nil
}

func replaceNeoGoModules(gomod string, baseContainer *dagger.Container, neoRootDir *dagger.Directory) (*dagger.Container, error) {
	replacedDeps, err := ReplacedDepsFromGoMod(gomod)
	if err != nil {
		return nil, err
	}

	// Add any replaced dependencies
	for _, dep := range replacedDeps {
		// replace ../ from dep
		dep = strings.Replace(dep, "../", "", 1)
		baseContainer = baseContainer.WithDirectory("/go/src/github.com/dmajrekar/dagger-cache-monorepo/"+dep, neoRootDir.Directory(dep))
	}

	return baseContainer, nil
}

// ReplacedDepsFromGoMod takes a go.mod file and returns the list of
// replaced dependencies
func ReplacedDepsFromGoMod(goModFile string) ([]string, error) {
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
