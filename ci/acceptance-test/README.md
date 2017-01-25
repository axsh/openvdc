# OpenVDC Acceptance Test

![Test environment drawing](illustrations/drawing.svg)

This is the environment used on the OpenVDC CI to run the integration tests. To run this environment locally first make a file containing the following environment variables.

```
# The following two lines make up the yum repository from which we'll download OpenVDC packages to test
# "https://ci.openvdc.org/repos/${BRANCH}/${RELEASE_SUFFIX}/"
BRANCH="master"
RELEASE_SUFFIX="current"

# Set to "1" if you don't want to remove the docker container after running
REBUILD="0"
```

Let's save this file as `build.env`. Now kick off the `build_and_run_in_docker.sh` script, passing in that file as an argument.

```
./build_and_run_in_docker.sh build.env
```

That's it. This should build the environment and run the tests
