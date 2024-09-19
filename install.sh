#!/bin/bash

set -e
OS=$(uname -s)
ARCH=$(uname -m)

if [[ "$ARCH" == "x86_64" ]]; then
    ARCH="amd64"
elif [[ "$ARCH" == "arm64" ]] || [[ "$ARCH" == "aarch64" ]]; then
    ARCH="arm64"
fi

API_URL="https://api.github.com/repos/dymensionxyz/roller/releases/latest"
if [ "$ROLLER_RELEASE_TAG" = "" ]; then
  TGZ_URL=$(curl -s "$API_URL" \
      | grep "browser_download_url.*_${OS}_${ARCH}.tar.gz" \
      | cut -d : -f 2,3 \
      | tr -d \" \
      | tr -d ' ' )
else
  TGZ_URL="https://github.com/dymensionxyz/roller/releases/download/$ROLLER_RELEASE_TAG/roller_${OS}_${ARCH}.tar.gz"
fi
ROLLER_BIN_PATH="/usr/local/bin/roller"
if [ -f "$ROLLER_BIN_PATH" ] || [ -f "$ROLLAPP_EVM_PATH" ] || [ -f "$DYMD_BIN_PATH" ] || [ -d "$INTERNAL_DIR" ]; then
    sudo rm -f "$ROLLER_BIN_PATH"
fi
sudo mkdir -p "/tmp/roller_tmp"
echo "💈 Downloading roller ${ROLLER_RELEASE_TAG}..."
sudo curl -L "$TGZ_URL" --progress-bar | sudo tar -xz -C "/tmp/roller_tmp"
echo "💈 Installing roller ${ROLLER_RELEASE_TAG}..."
sudo mv "/tmp/roller_tmp/roller" "$ROLLER_BIN_PATH"
sudo chmod +x "$ROLLER_BIN_PATH"
sudo rm -rf "/tmp/roller_tmp"
echo "💈 Installation complete! You can now use roller from your terminal."
