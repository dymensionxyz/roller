#!/bin/bash

# Abort if any command fails
set -e

# Emoji for echo outputs
EMOJI="💈"

# Detect the OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
VERSION="v0.1.2"

# The list of projects to be installed
PROJECTS=("dymension" "dymension-relayer" "roller" "dymension-rdk")
REPOS=("" "" "" "")
VERSIONS=("v1.0.0-rc2" "" "v0.1.0" "v0.4.0-rc2")
BUILDCOMMANDS=("" "" "" "make install_evm")
BINARYNAME=("dymd" "rly" "roller" "rollapp_evm")

# Set the appropriate download URL
if [[ "$ARCH" == "x86_64" ]]; then
    ARCH="amd64"
elif [[ "$ARCH" == "arm64" ]] || [[ "$ARCH" == "aarch64" ]]; then
    ARCH="arm64"
fi

TGZ_URL="https://github.com/dymensionxyz/roller/releases/download/${VERSION}/roller_${VERSION}_${OS}_${ARCH}.tar.gz"
# Create internal dir
INTERNAL_DIR="/usr/local/bin/roller_bins"
ROLLER_BIN_PATH="/usr/local/bin/roller"
ROLLAPP_EVM_PATH="/usr/local/bin/rollapp_evm"  # The path where rollapp_evm will be installed

# Check if Roller and rollapp_evm binaries already exist or the internal directory exists
if [ -f "$ROLLER_BIN_PATH" ] || [ -f "$ROLLAPP_EVM_PATH" ] || [ -d "$INTERNAL_DIR" ]; then
    sudo rm -f "$ROLLER_BIN_PATH"
    sudo rm -f "$ROLLAPP_EVM_PATH"
    sudo rm -rf "$INTERNAL_DIR"
fi

# Creating the required directories
sudo mkdir -p "$INTERNAL_DIR"
sudo mkdir -p "/tmp/roller_tmp"

# Function to display Go installation instructions
GO_REQUIRED_VERSION="1.19"
display_go_installation_instructions() {
    echo "$EMOJI To install Go $GO_REQUIRED_VERSION, you can run the following commands:"
    echo "  wget https://go.dev/dl/go1.19.10.${OS}-${ARCH}.tar.gz"
    echo "  sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.19.10.${OS}-${ARCH}.tar.gz"
}

# Download and extract the tar file to a temporary directory
echo "$EMOJI Downloading roller..."
if sudo curl -L "$TGZ_URL" --progress-bar | sudo tar -xz -C "/tmp/roller_tmp"; then
    # Assuming that the tar file contains the lib folder and the roller and rollapp_evm binaries inside the roller_bins directory.
    # Move binaries to their correct locations
    echo "$EMOJI Installing roller..."
    sudo mv "/tmp/roller_tmp/roller_bins/lib"/* "$INTERNAL_DIR"
    sudo mv "/tmp/roller_tmp/roller_bins/roller" "$ROLLER_BIN_PATH"
    sudo mv "/tmp/roller_tmp/roller_bins/rollapp_evm" "$ROLLAPP_EVM_PATH"  # move the rollapp_evm binary
else
    # Curl failed. Fallback to clone and build.
    echo "$EMOJI Download failed. Cloning and building manually..."
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
        if [ ! -d "/tmp/${PROJECTS[i]}" ]; then
            git clone "$REPO_URL"
        fi
        cd "${PROJECTS[i]}"
        git fetch
        if [ -n "${VERSIONS[i]}" ]; then
            git checkout "${VERSIONS[i]}"
        fi
        if [ -n "${BUILDCOMMANDS[i]}" ]; then
            ${BUILDCOMMANDS[i]}
        else
            make install
        fi

        for binary in ${BINARYNAME[i]}; do
            if [ ! -x "$(command -v "$binary")" ]; then
                echo "$EMOJI Couldn't find $binary in PATH. Aborting."; exit 1;
            fi
            binary_path=$(which "$binary" 2> /dev/null)  # Redirect error output to /dev/null
            sudo cp "$binary_path" "$INTERNAL_DIR"
        done
    done


    # Move roller and rollapp_evm separately
    sudo mv "$INTERNAL_DIR/roller" "$ROLLER_BIN_PATH"
    sudo mv "$INTERNAL_DIR/rollapp_evm" "$ROLLAPP_EVM_PATH"

fi



# Make roller and rollapp_evm executables
sudo chmod +x "$ROLLER_BIN_PATH"
sudo chmod +x "$ROLLAPP_EVM_PATH"

# Cleanup temporary directory
sudo rm -rf "/tmp/roller_tmp"

echo "$EMOJI Installation complete! You can now use roller from your terminal."
