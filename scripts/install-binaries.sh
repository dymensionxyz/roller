#!/bin/bash
BINS_DIR="/usr/local/bin"
ROLLER_BINS_DIR="$BINS_DIR/roller_bins"

ROLLAPP_EVN_VERSION="v2.2.1-rc01"
DYMD_VERSION="main"
DYMD_COMMIT="b9d863e6"
EIBC_VERSION="main"
RLY_VERSION="v0.3.4-v2.5.2-relayer"
CELESTIA_VERSION="v0.14.1"
CELESTIA_APP_VERSION="v1.11.0"

if [ -z "$BECH32_PREFIX" ]; then
  echo "please provide BECH32_PREFIX of the RollApp before running this script"
  exit 1
fi

# /usr/local/bin
# roller
if ! command -v "$BINS_DIR/roller" &> /dev/null; then
  cd ~/ && rm -rf roller/
  git clone https://github.com/dymensionxyz/roller.git && cd roller || exit 1
  make build && sudo mv ./build/roller "$BINS_DIR"
  cd ../ && rm -rf roller/
fi

# rollapp-evm
if ! command -v "$BINS_DIR/rollapp-evm" &> /dev/null; then
  cd ~/ && rm -rf rollapp-evm/
  git clone https://github.com/dymensionxyz/rollapp-evm.git --branch "$ROLLAPP_EVN_VERSION" && cd rollapp-evm || exit 1
  make build && sudo mv ./build/rollapp-evm "$BINS_DIR"
  cd ../ && rm -rf rollapp-evm/
fi

#dymd
if ! command -v "$BINS_DIR/dymd" &> /dev/null; then
  cd ~/ && rm -rf dymension/
  if [ "$DYMD_VERSION" != "" ]; then
    git clone https://github.com/dymensionxyz/dymension.git --branch "$DYMD_VERSION" && cd dymension || exit 1
    make build && sudo mv ./build/dymd "$BINS_DIR"
    cd ../ && rm -rf dymension/
  elif [ "$DYMD_COMMIT" != "" ]; then
    git clone https://github.com/dymensionxyz/dymension.git && cd dymension || exit 1
    git checkout $DYMD_COMMIT
    make build && sudo mv ./build/dymd "$BINS_DIR"
    cd ../ && rm -rf dymension/
  else
    exit 1
  fi
fi

# eibc
if ! command -v "$BINS_DIR/eibc-client" &> /dev/null; then
  cd ~/ && rm -rf eibc-client/
  git clone https://github.com/dymensionxyz/eibc-client.git --branch $EIBC_VERSION && cd eibc-client || exit 1
  make build && sudo mv ./build/eibc-client "$BINS_DIR"
  cd ../ && rm -rf eibc-client/
fi


# /usr/local/bin/roller_bins
if [ ! -d "$ROLLER_BINS_DIR" ]; then
    sudo mkdir -p "$ROLLER_BINS_DIR"
fi

# celestia & cel-key
if ! command -v "$ROLLER_BINS_DIR/celestia" &> /dev/null || ! command -v "$ROLLER_BINS_DIR/cel-key" &> /dev/null; then
  cd ~/ && rm -rf celestia-node
  git clone https://github.com/celestiaorg/celestia-node.git --branch $CELESTIA_VERSION && cd celestia-node || exit 1
  if ! command -v "$ROLLER_BINS_DIR/celestia" &> /dev/null; then
    make build && sudo mv build/celestia "$ROLLER_BINS_DIR"
  fi

  if ! command -v "$ROLLER_BINS_DIR/cel-key" &> /dev/null; then
    make cel-key && sudo mv cel-key "$ROLLER_BINS_DIR"
  fi
  cd ../ && rm -rf celestia-node
fi

# celestia-app
if ! command -v "$ROLLER_BINS_DIR/celestia-app" &> /dev/null; then
  cd ~/ && rm -rf celestia-app/
  git clone https://github.com/celestiaorg/celestia-app.git --branch $CELESTIA_APP_VERSION && cd celestia-app || exit 1
  make build && sudo mv build/celestia-app "$ROLLER_BINS_DIR"
  cd ../ && rm -rf celestia-app/
fi

# rly
if ! command -v "$ROLLER_BINS_DIR/rly" &> /dev/null; then
  cd ~/ && rm -rf go-relayer/
  git clone https://github.com/dymensionxyz/go-relayer.git --branch $RLY_VERSION && cd go-relayer || exit 1
  make build && sudo mv build/rly "$ROLLER_BINS_DIR"
  cd ../ && rm -rf go-relayer/
fi
