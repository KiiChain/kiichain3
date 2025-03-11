#!/bin/bash
set -e

teardown() {
    echo "Cleaning up resources..."
    make docker-cluster-stop-integration
}

# If any line fails, it will tear down docker
trap teardown EXIT

# Start docker environment
echo "Starting the docker environment"
sudo rm -f build/generated/launch.complete
sudo make docker-cluster-start-integration > /dev/null 2>&1 &

# Wait for liveness
until [ $(cat build/generated/launch.complete |wc -l) = 4 ]
do
    echo "Still initializing containers, sleeping..."
    sleep 1
done
echo "Nodes have started successfully. Sleeping for 30 seconds..."
sleep 30

echo "Starting the test..."
python3 integration_test/scripts/runner.py integration_test/upgrade/param_change_proposal.yaml