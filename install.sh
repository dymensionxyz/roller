#!/bin/bash

# Abort if any command fails
set -e

# Emoji for echo outputs
EMOJI="💈"

# Detect the OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
VERSION="v0.0.0"

# The list of projects to be installed
PROJECTS=("dymension" "dymension-relayer" "roller" "dymension-rdk" "celestia-node")
REPOS=("" "" "" "" "https://github.com/celestiaorg/celestia-node")
VERSIONS=("" "" "" "" "v0.6.4")
BUILDCOMMANDS=("" "" "" "make install_evm" "make go-install install-key")
BINARYNAME=("dymd" "rly" "roller" "rollapp_evm" "celestia cel-key")

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


# Check prerequisites
for cmd in git curl make; do
  if ! command -v $cmd >/dev/null 2>&1; then
    echo "Error: $cmd is not installed." >&2
    exit 1
  fi
done


# Check if Roller and rollapp_evm binaries already exist or the internal directory exists
if [ -f "$ROLLER_BIN_PATH" ] || [ -f "$ROLLAPP_EVM_PATH" ] || [ -d "$INTERNAL_DIR" ]; then
    sudo rm -f "$ROLLER_BIN_PATH"
    sudo rm -f "$ROLLAPP_EVM_PATH"
    sudo rm -rf "$INTERNAL_DIR"
fi

# Creating the required directories
sudo mkdir -p "$INTERNAL_DIR"
sudo mkdir -p "/tmp/roller_tmp"

# Download and extract the tar file to a temporary directory
echo "$EMOJI Downloading roller..."
if ! sudo curl -L "$TGZ_URL" --progress-bar | sudo tar -xz -C "/tmp/roller_tmp"; then
    # Curl failed. Fallback to clone and build.
    echo "$EMOJI Download failed. Cloning and building manually..."
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
        if [ -n "${VERSIONS[i]}" ]; then
            git checkout "${VERSIONS[i]}"
        fi
        if [ -n "${BUILDCOMMANDS[i]}" ]; then
            ${BUILDCOMMANDS[i]}
        else
            make install
        fi

        for binary in ${BINARYNAME[i]}; do
            binary_path=$(which "$binary" 2> /dev/null)  # Redirect error output to /dev/null
            sudo cp "$binary_path" "$INTERNAL_DIR"
        done
    done

    # Move roller and rollapp_evm separately
    sudo mv "$INTERNAL_DIR/roller" "$ROLLER_BIN_PATH"
    sudo mv "$INTERNAL_DIR/rollapp_evm" "$ROLLAPP_EVM_PATH"
else
    # Assuming that the tar file contains the lib folder and the roller and rollapp_evm binaries inside the roller_bins directory.
    # Move binaries to their correct locations
    echo "$EMOJI Installing roller..."
    sudo mv "/tmp/roller_tmp/roller_bins/lib"/* "$INTERNAL_DIR"
    sudo mv "/tmp/roller_tmp/roller_bins/roller" "$ROLLER_BIN_PATH"
    sudo mv "/tmp/roller_tmp/roller_bins/rollapp_evm" "$ROLLAPP_EVM_PATH"  # move the rollapp_evm binary
fi



# Make roller and rollapp_evm executables
sudo chmod +x "$ROLLER_BIN_PATH"
sudo chmod +x "$ROLLAPP_EVM_PATH"

# Cleanup temporary directory
sudo rm -rf "/tmp/roller_tmp"

echo "$EMOJI Installation complete! You can now use roller from your terminal."
