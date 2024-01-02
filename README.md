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

First, make sure you have the latest version of all the dependencies:

```bash
curl -L https://dymensionxyz.github.io/roller/install.sh | bash
```

To run the all the tests, run from the root directory:

```bash
go test ./...
```

## Installing a Pre Release

To install a specific pre-release version, use:

```bash
export ROLLER_RELEASE_TAG="<RELEASE_TAG>"
curl -L https://dymensionxyz.github.io/roller/install.sh | bash
```

For more information about Roller and its usage, please refer to [the documentation](https://docs.dymension.xyz/build/roller).
