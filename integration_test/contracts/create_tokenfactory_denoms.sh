#!/bin/bash

kiichaindbin=$(which ~/go/bin/kiichaind | tr -d '"')
keyname=$(printf "12345678\n" | $kiichaindbin keys list --output json | jq ".[0].name" | tr -d '"')
keyaddress=$(printf "12345678\n" | $kiichaindbin keys list --output json | jq ".[0].address" | tr -d '"')
chainid=$($kiichaindbin status | jq ".NodeInfo.network" | tr -d '"')
kiihome=$(git rev-parse --show-toplevel | tr -d '"')

cd $kiihome || exit
echo "Deploying first set of tokenfactory denoms..."

beginning_block_height=$($kiichaindbin status | jq -r '.SyncInfo.latest_block_height')
echo "$beginning_block_height" > $kiihome/integration_test/contracts/tfk_beginning_block_height.txt
echo "$keyaddress"  > $kiihome/integration_test/contracts/tfk_creator_id.txt

# create first set of tokenfactory denoms
for i in {1..10}
do
    echo "Creating first set of tokenfactory denoms #$i..."
    create_denom_result=$(printf "12345678\n" | $kiichaindbin tx tokenfactory create-denom "$i" -y --from="$keyname" --chain-id="$chainid" --gas=500000 --fees=100000ukii --broadcast-mode=block --output=json)
    new_token_denom=$(echo "$create_denom_result" | jq -r '.logs[].events[].attributes[] | select(.key == "new_token_denom").value')
    echo "Got token $new_token_denom for iteration $i"
done


first_set_block_height=$($kiichaindbin status | jq -r '.SyncInfo.latest_block_height')
echo "$first_set_block_height" > $kiihome/integration_test/contracts/tfk_first_set_block_height.txt

sleep 5

# create second set of tokenfactory denoms
for i in {11..20}
do
    echo "Creating first set of tokenfactory denoms #$i..."
    create_denom_result=$(printf "12345678\n" | $kiichaindbin tx tokenfactory create-denom "$i" -y --from="$keyname" --chain-id="$chainid" --gas=500000 --fees=100000ukii --broadcast-mode=block --output=json)
    new_token_denom=$(echo "$create_denom_result" | jq -r '.logs[].events[].attributes[] | select(.key == "new_token_denom").value')
    echo "Got token $new_token_denom for iteration $i"
done

second_set_block_height=$($kiichaindbin status | jq -r '.SyncInfo.latest_block_height')
echo "$second_set_block_height" > $kiihome/integration_test/contracts/tfk_second_set_block_height.txt

sleep 5

# create third set of tokenfactory denoms
for i in {21..30}
do
    echo "Creating first set of tokenfactory denoms #$i..."
    create_denom_result=$(printf "12345678\n" | $kiichaindbin tx tokenfactory create-denom "$i" -y --from="$keyname" --chain-id="$chainid" --gas=500000 --fees=100000ukii --broadcast-mode=block --output=json)
    new_token_denom=$(echo "$create_denom_result" | jq -r '.logs[].events[].attributes[] | select(.key == "new_token_denom").value')
    echo "Got token $new_token_denom for iteration $i"
done

third_set_block_height=$($kiichaindbin status | jq -r '.SyncInfo.latest_block_height')
echo "$third_set_block_height" > $kiihome/integration_test/contracts/tfk_third_set_block_height.txt

num_denoms=$(kiichaind q tokenfactory denoms-from-creator $CREATOR_ID --output json | jq -r ".denoms | length")
echo $num_denoms

exit 0
