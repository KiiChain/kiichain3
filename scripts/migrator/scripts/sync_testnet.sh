NODE_HOME=~/.kiichain3
NODE_MONIKER=testnet_oro
SERVICE_NAME=kiichain3
SERVICE_VERSION="v3.0.0"

# Binary
CHAIN_BINARY='kiichaind'
CHAIN_ID=kiichain3

# Persistent peers and RPC endpoints
PERSISTENT_PEERS="5b6aa55124c0fd28e47d7da091a69973964a9fe1@uno.sentry.testnet.v3.kiivalidator.com:26656,5e6b283c8879e8d1b0866bda20949f9886aff967@dos.sentry.testnet.v3.kiivalidator.com:26656"
PRIMARY_ENDPOINT=https://rpc.uno.sentry.testnet.v3.kiivalidator.com
SECONDARY_ENDPOINT=https://rpc.dos.sentry.testnet.v3.kiivalidator.com

# The genesis for the chain
GENESIS_URL=https://raw.githubusercontent.com/KiiChain/testnets/refs/heads/main/testnet_oro/genesis.json

# Initialize home directory
echo "Initializing $NODE_HOME..."
cd $HOME
rm -rf $NODE_HOME
$CHAIN_BINARY config chain-id $CHAIN_ID --home $NODE_HOME
$CHAIN_BINARY config keyring-backend test --home $NODE_HOME
$CHAIN_BINARY config broadcast-mode block --home $NODE_HOME
$CHAIN_BINARY init $NODE_MONIKER --chain-id $CHAIN_ID --home $NODE_HOME

# Set the PERSISTENT_PEERS
sed -i -e "/persistent-peers =/ s^= .*^= \"$PERSISTENT_PEERS\"^" $NODE_HOME/config/config.toml

# Configure state-sync
TRUST_HEIGHT_DELTA=500
LATEST_HEIGHT=$(curl -s "$PRIMARY_ENDPOINT"/block | jq -r ".block.header.height")
if [[ "$LATEST_HEIGHT" -gt "$TRUST_HEIGHT_DELTA" ]]; then
SYNC_BLOCK_HEIGHT=$(($LATEST_HEIGHT - $TRUST_HEIGHT_DELTA))
else
SYNC_BLOCK_HEIGHT=$LATEST_HEIGHT
fi

# Get the sync block hash
SYNC_BLOCK_HASH=$(curl -s "$PRIMARY_ENDPOINT/block?height=$SYNC_BLOCK_HEIGHT" | jq -r ".block_id.hash")

# Enable state sync
sed -i.bak -e "s|^enable *=.*|enable = true|" $NODE_HOME/config/config.toml
sed -i.bak -e "s|^rpc-servers *=.*|rpc-servers = \"$PRIMARY_ENDPOINT,$SECONDARY_ENDPOINT\"|" $NODE_HOME/config/config.toml
sed -i.bak -e "s|^db-sync-enable *=.*|db-sync-enable = false|" $NODE_HOME/config/config.toml
sed -i.bak -e "s|^trust-height *=.*|trust-height = $SYNC_BLOCK_HEIGHT|" $NODE_HOME/config/config.toml
sed -i.bak -e "s|^trust-hash *=.*|trust-hash = \"$SYNC_BLOCK_HASH\"|" $NODE_HOME/config/config.toml

# Set the node as validator
sed -i 's/mode = "full"/mode = "validator"/g' $NODE_HOME/config/config.toml

# Enable DB
sed -i.bak -e "s|^occ-enabled *=.*|occ-enabled = true|" $NODE_HOME/config/app.toml
sed -i.bak -e "s|^sc-enable *=.*|sc-enable = true|" $NODE_HOME/config/app.toml
sed -i.bak -e "s|^ss-enable *=.*|ss-enable = true|" $NODE_HOME/config/app.toml
sed -i.bak -e 's/^# concurrency-workers = 20$/concurrency-workers = 500/' $NODE_HOME/config/app.toml

# Replace genesis file
echo "Replacing genesis file..."
wget $GENESIS_URL -O genesis.json
mv genesis.json $NODE_HOME/config/genesis.json

# Replace keys
kiichaind start