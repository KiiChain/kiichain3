#!/usr/bin/env sh

NODE_ID=${ID:-0}
INVARIANT_CHECK_INTERVAL=${INVARIANT_CHECK_INTERVAL:-0}

LOG_DIR="build/generated/logs"
mkdir -p $LOG_DIR

echo "Starting the kiichaind process for node $NODE_ID with invariant check interval=$INVARIANT_CHECK_INTERVAL..."

cosmovisor run start --chain-id kiichain3 --inv-check-period ${INVARIANT_CHECK_INTERVAL} > "$LOG_DIR/kiichaind-$NODE_ID.log" 2>&1 &

echo "Node $NODE_ID kiichaind is started now"
echo "Done" >> build/generated/launch.complete
