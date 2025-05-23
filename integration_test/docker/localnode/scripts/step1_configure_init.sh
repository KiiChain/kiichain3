#!/usr/bin/env sh

# Input parameters
NODE_ID=${ID:-0}
NUM_ACCOUNTS=${NUM_ACCOUNTS:-5}
echo "Configure and initialize environment"

cp build/kiichaind "$GOBIN"/
# cp build/price-feeder "$GOBIN"/ TODO: Re-enable me after we have the oracle module again

# Prepare shared folders
mkdir -p build/generated/gentx/
mkdir -p build/generated/exported_keys/
mkdir -p build/generated/node_"$NODE_ID"

# Testing whether kiichaind works or not
kiichaind version # Uncomment the below line if there are any dependency issues
# ldd build/kiichaind

# Initialize validator node
MONIKER="kii-node-$NODE_ID"

kiichaind init "$MONIKER" --chain-id kii >/dev/null 2>&1

# Copy configs
ORACLE_CONFIG_FILE="build/generated/node_$NODE_ID/price_feeder_config.toml"
APP_CONFIG_FILE="build/generated/node_$NODE_ID/app.toml"
TENDERMINT_CONFIG_FILE="build/generated/node_$NODE_ID/config.toml"
cp integration_test/docker/localnode/config/app.toml "$APP_CONFIG_FILE"
cp integration_test/docker/localnode/config/config.toml "$TENDERMINT_CONFIG_FILE"
# cp integration_test/docker/localnode/config/price_feeder_config.toml "$ORACLE_CONFIG_FILE" TODO: Re-enable me after we have the oracle module again


# Set up persistent peers
KII_NODE_ID=$(kiichaind tendermint show-node-id)
NODE_IP=$(hostname -i | awk '{print $1}')
echo "$KII_NODE_ID@$NODE_IP:26656" >> build/generated/persistent_peers.txt

# Create a new account
ACCOUNT_NAME="node_admin"
echo "Adding account $ACCOUNT_NAME"
printf "12345678\n12345678\ny\n" | kiichaind keys add "$ACCOUNT_NAME" >/dev/null 2>&1

# Get genesis account info
GENESIS_ACCOUNT_ADDRESS=$(printf "12345678\n" | kiichaind keys show "$ACCOUNT_NAME" -a)
echo "$GENESIS_ACCOUNT_ADDRESS" >> build/generated/genesis_accounts.txt

# Add funds to genesis account
kiichaind add-genesis-account "$GENESIS_ACCOUNT_ADDRESS" 10000000ukii,10000000uusdc,10000000uatom

# Create gentx
printf "12345678\n" | kiichaind gentx "$ACCOUNT_NAME" 10000000ukii --chain-id kii
cp ~/.kiichain3/config/gentx/* build/generated/gentx/

# Creating some testing accounts
echo "Creating $NUM_ACCOUNTS accounts"
python3 loadtest/scripts/populate_genesis_accounts.py "$NUM_ACCOUNTS" loc >/dev/null 2>&1
echo "Finished $NUM_ACCOUNTS accounts creation"

# Set node kiivaloper info
KIIVALOPER_INFO=$(printf "12345678\n" | kiichaind keys show "$ACCOUNT_NAME" --bech=val -a)
PRIV_KEY=$(printf "12345678\n12345678\n" | kiichaind keys export "$ACCOUNT_NAME")
echo "$PRIV_KEY" >> build/generated/exported_keys/"$KIIVALOPER_INFO".txt

# Update price_feeder_config.toml with address info
# TODO: Re-enable me after we have the oracle module again
# sed -i.bak -e "s|^address *=.*|address = \"$GENESIS_ACCOUNT_ADDRESS\"|" $ORACLE_CONFIG_FILE
# sed -i.bak -e "s|^validator *=.*|validator = \"$KIIVALOPER_INFO\"|" $ORACLE_CONFIG_FILE

echo "DONE" >> build/generated/init.complete
