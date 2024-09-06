#!/bin/bash
BINS_DIR="/usr/local/bin"
ROLLER_BINS_DIR="$BINS_DIR/roller_bins"

CELESTIA_VERSION="v0.16.0-rc0"
CELESTIA_APP_VERSION="v2.0.0"

install_or_update() {
    local tool=$1
    local current_version=$2
    local target_version=$3
    local repo_url=$4
    local build_cmd=$5
    local bin_name=$6
    local bin_dir=$7
    local extra_cmd=$8

    if [ "$current_version" != "$target_version" ]; then
        echo "Updating $tool from version $current_version to $target_version"
        cd ~/ && rm -rf "$tool"/
        git clone -q "$repo_url" --branch "$target_version" && cd "$tool" || exit 1
        eval "$extra_cmd"
        eval "$build_cmd" && sudo mv "$bin_name" "$bin_dir"
        cd ../ && rm -rf "$tool"/
    else
        echo "$tool is already at the correct version $target_version"
    fi
}

# celestia & cel-key
current_version=$("$ROLLER_BINS_DIR/celestia" version 2>/dev/null | awk '/Semantic version:/ {print $3}' || echo "")
if [ "$current_version" != "$CELESTIA_VERSION" ] || ! command -v "$ROLLER_BINS_DIR/cel-key" &> /dev/null; then
    echo "Installing/Updating Celestia from version $current_version to $CELESTIA_VERSION"
    cd ~/ && rm -rf celestia-node
    git clone https://github.com/celestiaorg/celestia-node.git --branch "$CELESTIA_VERSION" && cd celestia-node || exit 1
    make build && sudo mv build/celestia "$ROLLER_BINS_DIR"
    make cel-key && sudo mv cel-key "$ROLLER_BINS_DIR"
    cd ../ && rm -rf celestia-node
else
    echo "Celestia is already at the correct version $CELESTIA_VERSION"
fi

# celestia-app
current_version=$("$ROLLER_BINS_DIR/celestia-appd" version)
install_or_update "celestia-app" "v$current_version" "$CELESTIA_APP_VERSION" \
    "https://github.com/celestiaorg/celestia-app.git" "make build" "build/celestia-appd" "$ROLLER_BINS_DIR"
