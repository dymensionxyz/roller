#!/bin/bash
random_number=$(( RANDOM % 100000 + 1 ))
# will be passed in the orchestrator script
IDENTIFIER="RAX"
/usr/local/bin/roller config init loadtest_${random_number}-1 $IDENTIFIER --hub froopyland --da local --no-output
HUB_SEQ_ADDR=$(/usr/local/bin/roller keys list --output json | jq -r '.hub_sequencer')
RELAYER_HUB_ADDR=$(/usr/local/bin/roller keys list --output json | jq -r '."relayer-hub-key"')
echo "HUB_SEQ_ADDR: $HUB_SEQ_ADDR"
echo "RELAYER_HUB_ADDR: $RELAYER_HUB_ADDR"
/usr/local/bin/dymd tx bank multi-send local-user $HUB_SEQ_ADDR $RELAYER_HUB_ADDR 20000000000000000000udym --yes -b block --keyring-backend test --node https://froopyland.rpc.silknodes.io:443 --fees 50000udym --chain-id froopyland_100-1
/usr/local/bin/roller register --no-output
