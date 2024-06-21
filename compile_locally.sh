#!/bin/bash

# Abort if any command fails
set -e

# Emoji for echo outputs
EMOJI="💈"

# Detect the OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# The list of projects to be installed
PROJECTS=("dymension" "go-relayer" "rollapp-evm")
REPOS=("" "" "")
VERSIONS=("v3.1.0" "v0.3.3-v2.5.2-relayer" "v2.2.0")
BUILDCOMMANDS=("" "" "")
BINARYNAME=("dymd" "rly" "rollapp-evm")

if [[ "$ARCH" == "x86_64" ]]; then
    ARCH="amd64"
elif [[ "$ARCH" == "arm64" ]] || [[ "$ARCH" == "aarch64" ]]; then
    ARCH="arm64"
fi

INTERNAL_DIR="/usr/local/bin/roller_bins"
ROLLER_BIN_PATH="/usr/local/bin/roller"
ROLLAPP_EVM_PATH="/usr/local/bin/rollapp_evm"
DYMD_PATH="/usr/local/bin/dymd"

if [ -f "$ROLLER_BIN_PATH" ] || [ -f "$DYMD_PATH" ] || [ -f "$ROLLAPP_EVM_PATH" ] || [ -d "$INTERNAL_DIR" ]; then
    sudo rm -f "$ROLLER_BIN_PATH"
    sudo rm -f "$ROLLAPP_EVM_PATH"
    sudo rm -f "$DYMD_PATH"
    sudo rm -rf "$INTERNAL_DIR"
fi

# Creating the required directories
sudo mkdir -p "$INTERNAL_DIR"
sudo mkdir -p "/tmp/roller_tmp"

# Function to display Go installation instructions
GO_REQUIRED_VERSION="1.20"
display_go_installation_instructions() {
    echo "$EMOJI To install Go $GO_REQUIRED_VERSION, you can run the following commands:"
    echo "  wget https://go.dev/dl/go1.20.13.${OS}-${ARCH}.tar.gz"
    echo "  sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.20.13.${OS}-${ARCH}.tar.gz"
}

# Check prerequisites
command -v curl >/dev/null 2>&1 || { echo >&2 "$EMOJI curl is required but it's not installed. Aborting."; exit 1; }
command -v wget >/dev/null 2>&1 || { echo >&2 "$EMOJI wget is required but it's not installed. Aborting."; exit 1; }
command -v git >/dev/null 2>&1 || { echo >&2 "$EMOJI git is required but it's not installed. Aborting."; exit 1; }
if ! command -v go >/dev/null 2>&1; then
    echo >&2 "$EMOJI Go is required but it's not installed. Aborting."
    display_go_installation_instructions
    exit 1;
fi

# Check go version
GO_VERSION=$(go version | awk -F' ' '{print substr($3, 3, 4)}')  # Extract major.minor version number (remove the 'go' prefix)
if [ "$GO_VERSION" != "$GO_REQUIRED_VERSION" ]; then
    echo "$EMOJI Go version $GO_REQUIRED_VERSION is required, but you have version $GO_VERSION. Aborting."
    display_go_installation_instructions
    exit 1
fi

for i in "${!PROJECTS[@]}"; do
    echo "$EMOJI handling ${PROJECTS[i]}..."
    cd /tmp

    # Check if binaries already exist
    IFS=' ' read -r -a binary_array <<< "${BINARYNAME[i]}"  # convert string to array
    binary=${binary_array[0]}  # get the first element
    echo "$EMOJI checking for $binary..."
    if [ -x "$(command -v "$binary")" ]; then
        read -p "$binary already exists. Do you want to overwrite it? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Nn]$ ]]; then
            binary_path=$(which "$binary" 2> /dev/null)  # Redirect error output to /dev/null
            sudo cp "$binary_path" "$INTERNAL_DIR"
            continue 1
        else
            rm -f "$binary_path"
        fi
    fi

    REPO_URL=${REPOS[i]:-"https://github.com/dymensionxyz/${PROJECTS[i]}"}
    rm -rf "/tmp/${PROJECTS[i]}"
    git clone "$REPO_URL"
    cd "${PROJECTS[i]}"
    if [ "${VERSIONS[i]}" != "" ]; then
        git checkout "${VERSIONS[i]}"
    fi
    if [ "${BUILDCOMMANDS[i]}" != "" ]; then
        "${BUILDCOMMANDS[i]}"
    else
        make install
    fi

    for binary in "${BINARYNAME[i]}"; do
        if [ ! -x "$(command -v "$binary")" ]; then
            echo "$EMOJI Couldn't find $binary in PATH. Aborting."; exit 1;
        fi
        binary_path=$(which "$binary" 2> /dev/null)  # Redirect error output to /dev/null
        sudo cp "$binary_path" "$INTERNAL_DIR"
    done
done

sudo mv "$INTERNAL_DIR/rollapp-evm" "$ROLLAPP_EVM_PATH"
sudo mv "$INTERNAL_DIR/dymd" "$DYMD_PATH"

sudo chmod +x "$ROLLAPP_EVM_PATH"
sudo chmod +x "$DYMD_PATH"

# Cleanup temporary directory
sudo rm -rf "/tmp/roller_tmp"

echo "$EMOJI Installation complete! You can now use roller from your terminal."
