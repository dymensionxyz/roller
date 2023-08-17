#!/bin/bash
set -e
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
    ARCH="amd64"
elif [[ "$ARCH" == "arm64" ]] || [[ "$ARCH" == "aarch64" ]]; then
    ARCH="arm64"
fi
API_URL="https://api.github.com/repos/dymensionxyz/roller/releases/latest"
if [ -z "$ROLLER_RELEASE_TAG" ]; then
  TGZ_URL=$(curl -s $API_URL \
      | grep "browser_download_url.*_${OS}_${ARCH}.tar.gz" \
      | cut -d : -f 2,3 \
      | tr -d \" \
      | tr -d ' ' )
else
  TGZ_URL="https://github.com/dymensionxyz/roller/releases/download/$ROLLER_RELEASE_TAG/roller_${ROLLER_RELEASE_TAG}_${OS}_${ARCH}.tar.gz"
fi
INTERNAL_DIR="/usr/local/bin/roller_bins"
ROLLER_BIN_PATH="/usr/local/bin/roller"
DYMD_BIN_PATH="/usr/local/bin/dymd"
ROLLAPP_EVM_PATH="/usr/local/bin/rollapp_evm"  # The path where rollapp_evm will be installed
if [ -f "$ROLLER_BIN_PATH" ] || [ -f "$ROLLAPP_EVM_PATH" ] || [ -f "$DYMD_BIN_PATH" ] || [ -d "$INTERNAL_DIR" ]; then
    sudo rm -f "$ROLLER_BIN_PATH"
    sudo rm -f "$ROLLAPP_EVM_PATH"
    sudo rm -f "$DYMD_BIN_PATH"
    sudo rm -rf "$INTERNAL_DIR"
fi
sudo mkdir -p "$INTERNAL_DIR"
sudo mkdir -p "/tmp/roller_tmp"
echo "💈 Downloading roller..."
sudo curl -L "$TGZ_URL" --progress-bar | sudo tar -xz -C "/tmp/roller_tmp"
echo "💈 Installing roller..."
sudo mv "/tmp/roller_tmp/roller_bins/lib"/* "$INTERNAL_DIR"
sudo mv "/tmp/roller_tmp/roller_bins/roller" "$ROLLER_BIN_PATH"
sudo mv "/tmp/roller_tmp/roller_bins/rollapp_evm" "$ROLLAPP_EVM_PATH"
sudo mv "/tmp/roller_tmp/roller_bins/dymd" "$DYMD_BIN_PATH"
sudo chmod +x "$ROLLER_BIN_PATH"
sudo chmod +x "$ROLLAPP_EVM_PATH"
sudo chmod +x "$DYMD_BIN_PATH"
sudo rm -rf "/tmp/roller_tmp"
echo "💈 Installation complete! You can now use roller from your terminal."
