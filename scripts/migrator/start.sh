#!/bin/bash

HOMEDIR="$HOME/.evmd"
MONIKER="localtestnet"
CHAINID="${CHAIN_ID:-kiichain_1336-1}"

LOGLEVEL="info"

# Clean the previous path
rm -rf "$HOMEDIR"

# Init the chain
evmd init $MONIKER -o --chain-id "$CHAINID" --home "$HOMEDIR"

# Copy the new json
cp /home/korok/kii/kiichain3/export_sorted.json $HOMEDIR/config/genesis.json

# Copy the private validator key from the other deployment
cp /home/korok/.kiichain3/config/priv_validator_key.json $HOMEDIR/config/priv_validator_key.json

# Start the node
evmd start "$TRACE" \
	--log_level $LOGLEVEL \
	--minimum-gas-prices=0.0001ukii \
	--home "$HOMEDIR" \
	--json-rpc.api eth,txpool,personal,net,debug,web3 \
	--chain-id "$CHAINID" \
    --json-rpc.enable true \
    --inv-check-period 1
