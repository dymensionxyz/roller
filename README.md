# Roller CLI

![Roller CLI Logo]("./images/readme.png")

## Introduction

Roller CLI is a command-line interface tool designed to facilitate the creation
and operation of RollApps on the Dymension Hub.
It allows developers to effortlessly spin up and run RollApps, simplifying the
development process and making the Dymension Hub
more accessible.

## Installation

To install Roller CLI, simply run the following command:

```bash
curl -L https://github.com/dymensionxyz/roller/releases/download/v0.0.0/install.sh
 | bash
```

This will automatically fetch the latest version of Roller CLI and install it on
 your
local machine.

## Usage

Currently, Roller CLI supports the following commands:

### `roller version`

This command shows the currently installed version details of Roller CLI.
To use it, type:

```bash
roller version
```

### `roller config init <rollappID> <denom>`

This command initializes rollapp configuration files on your local machine.
To use it, replace `<rollappID>` and `<denom>` with your rollapp ID and
denomination, respectively.
For instance:

```bash
roller config init myrollapp mydenom
```
