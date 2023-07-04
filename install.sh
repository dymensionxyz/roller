#!/bin/bash

# Abort if any command fails
set -e

# Emoji for echo outputs
EMOJI="💈"

# Detect the OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
VERSION="v0.0.0"

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

# Download and extract the tar file to a temporary directory
echo "$EMOJI Downloading roller..."
sudo curl -L "$TGZ_URL" --progress-bar | sudo tar -xz -C "/tmp/roller_tmp"

# Assuming that the tar file contains the lib folder and the roller and rollapp_evm binaries inside the roller_bins directory.
# Move binaries to their correct locations
echo "$EMOJI Installing roller..."
sudo mv "/tmp/roller_tmp/roller_bins/lib"/* "$INTERNAL_DIR"
sudo mv "/tmp/roller_tmp/roller_bins/roller" "$ROLLER_BIN_PATH"
sudo mv "/tmp/roller_tmp/roller_bins/rollapp_evm" "$ROLLAPP_EVM_PATH"  # move the rollapp_evm binary

# Make roller and rollapp_evm executables
sudo chmod +x "$ROLLER_BIN_PATH"
sudo chmod +x "$ROLLAPP_EVM_PATH"

# Cleanup temporary directory
sudo rm -rf "/tmp/roller_tmp"

echo "$EMOJI Installation complete! You can now use roller from your terminal."
