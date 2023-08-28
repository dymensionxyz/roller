#!/bin/bash

DEST_ADDR=${DEST_ADDR:-""}
ROLLAPP_ID=${ROLLAPP_ID:-""}
ROLLER_CONFIG_PATH=${ROLLER_CONFIG_PATH:-""}

if [ -z "$DEST_ADDR" ] || [ -z "$ROLLAPP_ID" ] || [ -z "$ROLLER_CONFIG_PATH" ]; then
    echo "Error: One or more required environment variables (DEST_ADDR, ROLLAPP_ID, ROLLER_CONFIG_PATH) are unset."
    exit 1
fi

TIMEOUT=1800 # 30 minutes

check_balance() {
    BALANCE=$(/usr/local/bin/dymd query bank balances $DEST_ADDR | grep "ibc/")
    if [[ $BALANCE ]]; then
        echo "IBC transaction relayed successfully!"
        exit 0
    fi
}

START_TIME=$(date +%s)

while true; do
    CURRENT_TIME=$(date +%s)
    ELAPSED_TIME=$(($CURRENT_TIME - $START_TIME))
    if [ $ELAPSED_TIME -ge $TIMEOUT ]; then
        echo "Timeout reached without successful IBC relay."
        exit 1
    fi
    /usr/local/bin/rollapp_evm tx ibc-transfer transfer transfer channel-0 "$DEST_ADDR" 1uRAX --from rollapp_sequencer --keyring-backend test --packet-timeout-timestamp 6000000000000 --broadcast-mode block -y --chain-id "$ROLLAPP_ID" --home "$ROLLER_CONFIG_PATH/rollapp"
    check_balance
    sleep 10
done
