#!/bin/bash

# Abort if any command fails
set -e

# Emoji for echo outputs
EMOJI="💈"

TGZ_URL="https://github.com/dymensionxyz/roller/releases/download/v0.0.0/roller_0.0.0_darwin_amd64.tar.gz"
# Create internal dir
INTERNAL_DIR="/usr/local/bin/roller_bins"
ROLLER_BIN_PATH="/usr/local/bin/roller"

# Check if Roller binary already exists
if [ -f "$ROLLER_BIN_PATH" ] || [ -d "$INTERNAL_DIR" ]; then
    sudo rm -f "$ROLLER_BIN_PATH"
    sudo rm -rf "$INTERNAL_DIR"
fi

# Creating the required directories
sudo mkdir -p "$INTERNAL_DIR"
sudo mkdir -p "/tmp/roller_tmp"

# Download and extract the tar file to a temporary directory
echo "$EMOJI Downloading roller..."
sudo curl -L "$TGZ_URL" --progress-bar | sudo tar -xz -C "/tmp/roller_tmp"

# Assuming that the tar file contains the lib folder and the roller binary inside the roller_bins directory.
# Move binaries to their correct locations
echo "$EMOJI Installing roller..."
sudo mv "/tmp/roller_tmp/roller_bins/lib"/* "$INTERNAL_DIR"
sudo mv "/tmp/roller_tmp/roller_bins/roller" "$ROLLER_BIN_PATH"

# Make roller executable
sudo chmod +x "$ROLLER_BIN_PATH"

# Cleanup temporary directory
sudo rm -rf "/tmp/roller_tmp"

echo "$EMOJI Installation complete! You can now use roller from your terminal."
