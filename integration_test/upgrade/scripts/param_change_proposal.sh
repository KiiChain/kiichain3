#!/bin/bash

PROPOSAL_ID=$(printf "12345678\n"  | kiichaind tx gov submit-proposal param-change integration_test/upgrade/scripts/proposal.json --from node_admin --fees 2000ukii -b block -y --output json | jq -M -r ".logs[].events[].attributes[0] | select(.key == \"proposal_id\").value")

echo $PROPOSAL_ID
