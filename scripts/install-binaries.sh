#!/bin/bash
BINS_DIR="/usr/local/bin"
ROLLER_BINS_DIR="$BINS_DIR/roller_bins"

ROLLAPP_EVN_VERSION="main"
DYMD_VERSION="main"
DYMD_COMMIT="f42674ae"

EIBC_VERSION="main"
RLY_VERSION="v0.3.4-v2.5.2-relayer"
CELESTIA_VERSION="v0.14.1"
CELESTIA_APP_VERSION="v1.11.0"

if [ "$BECH32_PREFIX" = "" ]; then
  echo "please provide BECH32_PREFIX of the RollApp before running this script"
  exit 1
fi

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

# roller
if ! command -v "$BINS_DIR/roller" &> /dev/null; then
    install_or_update "roller" "" "main" \
        "https://github.com/dymensionxyz/roller.git" "make buill" "./build/roller" "$BINS_DIR"
fi

# rollapp-evm
current_version=$("$BINS_DIR/rollapp-evm" version 2>/dev/null | cut -d ' ' -f 3 || echo "")
install_or_update "rollapp-evm" "$current_version" "$ROLLAPP_EVN_VERSION" \
    "https://github.com/dymensionxyz/rollapp-evm.git" "make build" "./build/rollapp-evm" "$BINS_DIR"

# dymd
current_version=$("$BINS_DIR/dymd" version 2>/dev/null | cut -d ' ' -f 3 || echo "")
if [ "$DYMD_VERSION" != "main" ]; then
    install_or_update "dymension" "$current_version" "$DYMD_VERSION" \
        "https://github.com/dymensionxyz/dymension.git" "make build" "./build/dymd" "$BINS_DIR"
else

    if [ "$DYMD_COMMIT" != "" ]; then
        echo "Installing dymd from main branch with specific commit $DYMD_COMMIT"
        cd ~/ && rm -rf dymension/
        git clone https://github.com/dymensionxyz/dymension.git && cd dymension || exit 1
        git checkout "$DYMD_COMMIT"
        make build && sudo mv ./build/dymd "$BINS_DIR"
        cd ../ && rm -rf dymension/
    else
        echo "Error: DYMD_VERSION is set to main and DYMD_COMMIT is not specified"
        exit 1
    fi
fi

# eibc
current_version=$("$BINS_DIR/eibc-client" version 2>/dev/null | cut -d ' ' -f 3 || echo "")
install_or_update "eibc-client" "$current_version" "$EIBC_VERSION" \
    "https://github.com/dymensionxyz/eibc-client.git" "make build" "./build/eibc-client" "$BINS_DIR"

# Create ROLLER_BINS_DIR if it doesn't exist
if [ ! -d "$ROLLER_BINS_DIR" ]; then
    sudo mkdir -p "$ROLLER_BINS_DIR"
fi

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

# rly
current_version=$("$ROLLER_BINS_DIR/rly" version 2>/dev/null | awk '/^version:/ {print $2}' || echo "")
if [ "$current_version" = "" ] || [ "$current_version" != "${RLY_VERSION#v}" ]; then
    install_or_update "go-relayer" "$current_version" "$RLY_VERSION" \
        "https://github.com/dymensionxyz/go-relayer.git" "make build" "build/rly" "$ROLLER_BINS_DIR"
else
    echo "rly is already at the correct version $RLY_VERSION"
fi
