#!/bin/bash
# require success for commands
set -e


# Use python3 as default, but fall back to python if python3 doesn't exist
PYTHON_CMD=python3
if ! command -v $PYTHON_CMD &> /dev/null
then
    PYTHON_CMD=python
fi

# myKey address 0x7cb61d4117ae31a12e393a1cfa3bac666481d02e | os10jmp6sgh4cc6zt3e8gw05wavvejgr5pwjnpcky
VAL_KEY="mykey"
VAL_MNEMONIC="gesture inject test cycle original hollow east ridge hen combine junk child bacon zero hope comfort vacuum milk pitch cage oppose unhappy lunar seat"

# dev0 address 0xc6fe5d33615a1c52c08018c47e8bc53646a0e101 | os1cml96vmptgw99syqrrz8az79xer2pcgp84pdun
USER1_KEY="dev0"
USER1_MNEMONIC="copper push brief egg scan entry inform record adjust fossil boss egg comic alien upon aspect dry avoid interest fury window hint race symptom"

# set key name
# Uncomment the following if you'd like to run jaeger
#docker stop jaeger
#docker rm jaeger
#docker run -d --name jaeger \
#  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
#  -p 5775:5775/udp \
#  -p 6831:6831/udp \
#  -p 6832:6832/udp \
#  -p 5778:5778 \
#  -p 16686:16686 \
#  -p 14250:14250 \
#  -p 14268:14268 \
#  -p 14269:14269 \
#  -p 9411:9411 \
#  jaegertracing/all-in-one:1.33
# clean up old kiichain directory

rm -rf ~/.kiichain3
echo "Building..."
#install kiichaind
make install
# initialize chain with chain ID and add the first key
kiichaind init demo --chain-id kiichain3
echo "$VAL_MNEMONIC" | kiichaind keys add $VAL_KEY --keyring-backend test --recover
echo "$USER1_MNEMONIC" | kiichaind keys add $USER1_KEY --keyring-backend test --recover
# add the key as a genesis account with massive balances of several different tokens
kiichaind add-genesis-account $(kiichaind keys show $VAL_KEY -a --keyring-backend test) 1000000000000000ukii --keyring-backend test
kiichaind add-genesis-account $(kiichaind keys show $USER1_KEY -a --keyring-backend test) 1000000000000000ukii --keyring-backend test
# gentx for account
kiichaind gentx $VAL_KEY 100000000000000ukii --chain-id kiichain3 --keyring-backend test
# add validator information to genesis file
KEY=$(jq '.pub_key' ~/.kiichain3/config/priv_validator_key.json -c)
jq '.validators = [{}]' ~/.kiichain3/config/genesis.json > ~/.kiichain3/config/tmp_genesis.json
jq '.validators[0] += {"power":"100000000"}' ~/.kiichain3/config/tmp_genesis.json > ~/.kiichain3/config/tmp_genesis_2.json
jq '.validators[0] += {"pub_key":'$KEY'}' ~/.kiichain3/config/tmp_genesis_2.json > ~/.kiichain3/config/tmp_genesis_3.json
mv ~/.kiichain3/config/tmp_genesis_3.json ~/.kiichain3/config/genesis.json && rm ~/.kiichain3/config/tmp_genesis.json && rm ~/.kiichain3/config/tmp_genesis_2.json

echo "Creating Accounts"
kiichaind collect-gentxs
# update some params in genesis file for easier use of the chain locals (make gov props faster)
cat ~/.kiichain3/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["max_deposit_period"]="60s"' > ~/.kiichain3/config/tmp_genesis.json && mv ~/.kiichain3/config/tmp_genesis.json ~/.kiichain3/config/genesis.json
cat ~/.kiichain3/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="30s"' > ~/.kiichain3/config/tmp_genesis.json && mv ~/.kiichain3/config/tmp_genesis.json ~/.kiichain3/config/genesis.json
cat ~/.kiichain3/config/genesis.json | jq '.app_state["gov"]["voting_params"]["expedited_voting_period"]="10s"' > ~/.kiichain3/config/tmp_genesis.json && mv ~/.kiichain3/config/tmp_genesis.json ~/.kiichain3/config/genesis.json
cat ~/.kiichain3/config/genesis.json | jq '.app_state["oracle"]["params"]["vote_period"]="2"' > ~/.kiichain3/config/tmp_genesis.json && mv ~/.kiichain3/config/tmp_genesis.json ~/.kiichain3/config/genesis.json
cat ~/.kiichain3/config/genesis.json | jq '.app_state["oracle"]["params"]["whitelist"]=[{"name": "ueth"},{"name": "ubtc"},{"name": "uusdc"},{"name": "uusdt"},{"name": "uosmo"},{"name": "uatom"},{"name": "ukii"}]' > ~/.kiichain3/config/tmp_genesis.json && mv ~/.kiichain3/config/tmp_genesis.json ~/.kiichain3/config/genesis.json
cat ~/.kiichain3/config/genesis.json | jq '.app_state["distribution"]["params"]["community_tax"]="0.000000000000000000"' > ~/.kiichain3/config/tmp_genesis.json && mv ~/.kiichain3/config/tmp_genesis.json ~/.kiichain3/config/genesis.json
cat ~/.kiichain3/config/genesis.json | jq '.consensus_params["block"]["max_gas"]="35000000"' > ~/.kiichain3/config/tmp_genesis.json && mv ~/.kiichain3/config/tmp_genesis.json ~/.kiichain3/config/genesis.json
cat ~/.kiichain3/config/genesis.json | jq '.app_state["staking"]["params"]["max_voting_power_ratio"]="1.000000000000000000"' > ~/.kiichain3/config/tmp_genesis.json && mv ~/.kiichain3/config/tmp_genesis.json ~/.kiichain3/config/genesis.json
cat ~/.kiichain3/config/genesis.json | jq '.app_state["bank"]["denom_metadata"]=[{"denom_units":[{"denom":"ukii","exponent":0,"aliases":["ukii"]}],"base":"ukii","display":"ukii","name":"ukii","symbol":"ukii"}]' > ~/.kiichain3/config/tmp_genesis.json && mv ~/.kiichain3/config/tmp_genesis.json ~/.kiichain3/config/genesis.json

# Use the Python command to get the dates
if [ ! -z "$2" ]; then
  APP_TOML_PATH="$2"
else
  APP_TOML_PATH="$HOME/.kiichain3/config/app.toml"
fi
# Enable OCC and KiichainDB
sed -i.bak -e 's/# concurrency-workers = .*/concurrency-workers = 500/' $APP_TOML_PATH
sed -i.bak -e 's/occ-enabled = .*/occ-enabled = true/' $APP_TOML_PATH
sed -i.bak -e 's/sc-enable = .*/sc-enable = true/' $APP_TOML_PATH
sed -i.bak -e 's/ss-enable = .*/ss-enable = true/' $APP_TOML_PATH


# set block time to 2s
if [ ! -z "$1" ]; then
  CONFIG_PATH="$1"
else
  CONFIG_PATH="$HOME/.kiichain3/config/config.toml"
fi

if [ ! -z "$2" ]; then
  APP_PATH="$2"
else
  APP_PATH="$HOME/.kiichain3/config/app.toml"
fi

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  sed -i 's/mode = "full"/mode = "validator"/g' $CONFIG_PATH
  sed -i 's/indexer = \["null"\]/indexer = \["kv"\]/g' $CONFIG_PATH
  sed -i 's/timeout_prevote =.*/timeout_prevote = "2000ms"/g' $CONFIG_PATH
  sed -i 's/timeout_precommit =.*/timeout_precommit = "2000ms"/g' $CONFIG_PATH
  sed -i 's/timeout_commit =.*/timeout_commit = "2000ms"/g' $CONFIG_PATH
  sed -i 's/skip_timeout_commit =.*/skip_timeout_commit = false/g' $CONFIG_PATH
  # sed -i 's/slow = false/slow = true/g' $APP_PATH
elif [[ "$OSTYPE" == "darwin"* ]]; then
  sed -i '' 's/mode = "full"/mode = "validator"/g' $CONFIG_PATH
  sed -i '' 's/indexer = \["null"\]/indexer = \["kv"\]/g' $CONFIG_PATH
  sed -i '' 's/unsafe-propose-timeout-override =.*/unsafe-propose-timeout-override = "2s"/g' $CONFIG_PATH
  sed -i '' 's/unsafe-propose-timeout-delta-override =.*/unsafe-propose-timeout-delta-override = "2s"/g' $CONFIG_PATH
  sed -i '' 's/unsafe-vote-timeout-override =.*/unsafe-vote-timeout-override = "2s"/g' $CONFIG_PATH
  sed -i '' 's/unsafe-vote-timeout-delta-override =.*/unsafe-vote-timeout-delta-override = "2s"/g' $CONFIG_PATH
  sed -i '' 's/unsafe-commit-timeout-override =.*/unsafe-commit-timeout-override = "2s"/g' $CONFIG_PATH
  # sed -i '' 's/slow = false/slow = true/g' $APP_PATH
else
  printf "Platform not supported, please ensure that the following values are set in your config.toml:\n"
  printf "###         Consensus Configuration Options         ###\n"
  printf "\t timeout_prevote = \"2000ms\"\n"
  printf "\t timeout_precommit = \"2000ms\"\n"
  printf "\t timeout_commit = \"2000ms\"\n"
  printf "\t skip_timeout_commit = false\n"
  exit 1
fi

kiichaind config keyring-backend test

if [ $NO_RUN = 1 ]; then
  echo "No run flag set, exiting without starting the chain"
  exit 0
fi

# start the chain with log tracing
GORACE="log_path=/tmp/race/kiichaind_race" kiichaind start --trace --chain-id kiichain3


kiichaind tx evm send 0xC6Fe5D33615a1C52c08018c47E8Bc53646A0E101 100000000000000000000000 --from mykey --keyring-backend test
