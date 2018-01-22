# Docker Automatic Build Tool #

## Requirements ##

This tool requires the `docker` binary be in the running user's path.

## Overview ##

The docker automatic build tool is designed to improve re-use and reduce duplication when creating docker images.

Images can be either base images - images that are used for infrastructure or to host code - or deployments - images that are packaged with a versioned, ready-to-run application.

The tool can be run either via command line or with web endpoints (using the `serve` command).

## Quick Start ##

To get started, get and install the build tool.

```
go get go.mikenewswanger.com/container-factory
go install go.mikenewswanger.com/container-factory
```

The tool can then be launched using `container-factory`

To see the underlying commands being run, commands can be run with verbosity level 3 `-vvv`.

## Build Commands #

A few examples are provided under the `/.examples/` directory.

To view the images that will be built during the automatic build process and their inheritance:
```
docker-automatic-build list-base-images -d $GOPATH/src/go.mikenewswanger.com/container-factory/.example
```

A `0` exit code indicates no issues building the hierarchy.

To build the base images:

```
docker-automatic-build build-base-images -d $GOPATH/src/go.mikenewswanger.com/container-factory/.example -p docker-registry.localhost --local-only
```

*Note*: Base images cannot be built individually.  If no changes were made to the underlying docker files, the docker agent will perform a no-op.  To force a rebuild, use the `--force-rebuild` option.

To see available deployments:
```
docker-automatic-build list-deployments -d $GOPATH/src/go.mikenewswanger.com/container-factory/.example
```

To build a deployment:
```
docker-automatic-build build-deployment -d $GOPATH/src/go.mikenewswanger.com/container-factory/.example -p docker-registry.localhost example --local-only
```

Deployments will always be prefixed with `/deployments/` in the tag.  The above image can be run as a container:
```
docker run --rm -ti docker-registry.localhost/deployments/example:<username> sh
```

See the output of `container-factory <subcommand> --help` for more details.

To push to a remote registry, remove `--local-only` from the above commands.

## Organizing Dockerfiles / Deployments ##

Dockerfiles and deployments will be tagged based on the folder structure in their respective directories.  If your registry supports it, you can nest images as deep as you'd like.
