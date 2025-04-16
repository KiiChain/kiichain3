#!/bin/bash
set -ue

#############
# VARIABLES #
#############

# Set the project variables
EXPORT_GENESIS="testnet_data/export_testnet.json"
MIGRATED_GENESIS="testnet_data/migrated_genesis.json"
# Validator #1 data
VAL_1_HOME="/tmp/kiichain/home_validator1"
VAL_1_NODE_KEY="testnet_data/node_key_1.json"
VAL_1_PRIV_VALIDATOR_KEY="testnet_data/priv_key_1.json"
VAL_1_P2P_PORT=26156
VAL_1_NODE_KEY="3155135d22707c4021808f788e2bed44119dc687"
# Validator #2 data
VAL_2_HOME="/tmp/kiichain/home_validator2"
VAL_2_NODE_KEY="testnet_data/node_key_2.json"
VAL_2_PRIV_VALIDATOR_KEY="testnet_data/priv_key_2.json"
VAL_2_P2P_PORT=26256
VAL_2_NODE_KEY="6873de3ec2d1dc35c0d900ed2760eedafefe58ab"
# Validator #3 data
VAL_3_HOME="/tmp/kiichain/home_validator3"
VAL_3_NODE_KEY="testnet_data/node_key_3.json"
VAL_3_PRIV_VALIDATOR_KEY="testnet_data/priv_key_3.json"
VAL_3_P2P_PORT=26356
VAL_3_NODE_KEY="efc85289d64d3bf351fce0598e2fe160ab8819cf"

# Persistent peers setup
PERSISTENT_PEERS="$VAL_1_NODE_KEY@localhost:$VAL_1_P2P_PORT,$VAL_2_NODE_KEY@localhost:$VAL_2_P2P_PORT,$VAL_3_NODE_KEY@localhost:$VAL_3_P2P_PORT"

# Clean the old initialization home
rm -rf /tmp/kiichain

#############
# MIGRATION #
#############

# Start by migrating the genesis
source .venv/bin/activate
python scripts/migrator/main.py $EXPORT_GENESIS $MIGRATED_GENESIS

# Validate the genesis
kiichaind genesis validate $MIGRATED_GENESIS

##################
# INITIALIZATION #
##################

# Generic initialization function
start_validator() {
    local home=$1
    local node_key=$2
    local priv_validator_key=$3
    local moniker=$4

    # Start the path
    mkdir -p $home
    kiichaind init $moniker --home $home

    # Copy the genesis
    cp $MIGRATED_GENESIS $home/config/genesis.json
    # Copy the node key
    cp $node_key $home/config/node_key.json
    # Copy the priv validator key
    cp $priv_validator_key $home/config/priv_validator_key.json

    # Setup the p2p
    sed -i -e "/persistent_peers =/ s^= .*^= \"$PERSISTENT_PEERS\"^" $home/config/config.toml
}

# Start the validator #1
start_validator $VAL_1_HOME $VAL_1_NODE_KEY $VAL_2_PRIV_VALIDATOR_KEY validator1
# Start the validator #2
start_validator $VAL_2_HOME $VAL_2_NODE_KEY $VAL_2_PRIV_VALIDATOR_KEY validator2
# Start the validator #3
start_validator $VAL_3_HOME $VAL_3_NODE_KEY $VAL_3_PRIV_VALIDATOR_KEY validator3
