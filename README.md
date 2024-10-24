Repo demonstrating ExperimentalPrivilegedNesting showing cache misses

To replicate this:

- checkout https://github.com/DMajrekar/dagger-cache-monorepo/tree/broken-cache
- Run the pipeline dagger -m ci call run-test
- Note that the run takes 20 seconds
- Run the pipeline again dagger -m ci call run-test
- note the run is instant
- Increase the value of project1/VERSION to simulate a code change
- Run the pipeline again dagger -m ci call run-test
- note that the run is 10 seconds
- Increase the value of project1/VERSION to simulate a code change
- Run the pipeline with ExperimentalPrivilegedNesting  enabled dagger -m ci call run-test --enable-experimental-privileged-nesting
- note that the run time is 20 seconds
- Run the pipeline with ExperimentalPrivilegedNesting  enabled dagger -m ci call run-test --enable-experimental-privileged-nesting
- Note that the run is instant
- Increment the project1/VERSION to simulate a code change
- note that the run time is now 20 seconds, not 10 seconds
