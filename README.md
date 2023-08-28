# Roller CLI

![Roller CLI Logo](images/readme.png)

## Introduction

Roller CLI is a command-line interface tool designed to facilitate the creation
and operation of RollApps on the Dymension Hub.
It allows developers to effortlessly spin up and run RollApps, simplifying the
development process and making the Dymension Hub
more accessible.

## Local Development

To build and test the latest version from the main branch:

First, install all the necessary dependencies using the following command:

```bash
curl -L https://dymensionxyz.github.io/roller/install.sh | bash
```

Next, clone this repository. Once cloned, navigate to the root directory and execute:

```bash
make build
```

This command builds the latest version of Roller and places the executable
in the `./build` directory.

To run Roller, use:

```bash
./build/roller
```

## Testing

To run the all the tests, run from the root directory:

```bash
go test ./...
```

## Creating a New Release

Roller features a convenient Continuous Deployment (CD) workflow that
automatically generates all required assets after a new release is
created. To use it, simply create a new release on GitHub, and the
process will build and upload the release assets for you.
(works also for pre-releases)
By default, when installing Roller with

```bash
curl -L https://dymensionxyz.github.io/roller/install.sh | bash
```

It will install the latest release.
To install a specific version, use:

```bash
export ROLLER_RELEASE_TAG="<RELEASE_TAG>"
curl -L https://dymensionxyz.github.io/roller/install.sh | bash
```

For more information about Roller and its usage, please refer to [the documentation](https://docs.dymension.xyz/build/roller).
