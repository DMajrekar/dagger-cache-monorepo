# Cache Misses with ExperimentalPrivilegedNesting

This repo demonstrates a cache miss issue when using the `ExperimentalPrivilegedNesting` feature in Dagger.

There are two sample mono-repo type services within the `project1` and `project2` directories. Each project loads a shared library from the `lib` dir, and runs a 10 second sleep when run. To simulate a more complex repo, the sleep command has been refactored into the `lib` dir.

The dagger module in CI has a single function; `run-test`. This will:

1. Load each project Dir
2. Read the go.mod file in the project dir
3. Create a base container from the version of go in the respective go.mod file
4. Find any replaced dependencies in the go.mod file, and add them to the base container
5. Run `main.go` in the project dir, which sleeps for 10 seconds

When running without the `ExperimentalPrivilegedNesting` feature, the first run of this pipeline, takes 20 seconds, and subsequent runs with no changes are "instant" as expected. When the `VERSION` file in the `project1` dir is incremented, and the pipeline re-run, the pipeline takes 10 seconds to run, as expected.

When running with the `ExperimentalPrivilegedNesting` feature, the first run of the pipeline takes 20 seconds, and subsequent runs with no changes are "instant" as expected. When the `VERSION` file in the `project1` dir is incremented, and the pipeline re-run, the pipeline takes 20 seconds to run, not 10 seconds as expected.

## Steps to reproduce

- checkout https://github.com/DMajrekar/dagger-cache-monorepo
- Run the pipeline `dagger -m ci call run-test`
- Note that the run takes 20 seconds
- Run the pipeline again `dagger -m ci call run-test`
- note the run is instant
- Increase the value of project1/VERSION to simulate a code change
- Run the pipeline again `dagger -m ci call run-test`
- note that the run is 10 seconds

  
- Increase the value of project1/VERSION to simulate a code change
- Run the pipeline with ExperimentalPrivilegedNesting  enabled `dagger -m ci call run-test --enable-experimental-privileged-nesting`
- note that the run time is 20 seconds
- Run the pipeline with ExperimentalPrivilegedNesting  enabled `dagger -m ci call run-test --enable-experimental-privileged-nesting`
- Note that the run is instant
- Increment the project1/VERSION to simulate a code change
- note that the run time is now 20 seconds, not 10 seconds
