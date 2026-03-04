#!/bin/sh
echo "================================================"
echo "HLABS TigerBeetle Initialization Script"
echo "================================================"

set -e # Stops the script if a critical error occurs

# Path to the data file inside the container volume
TB_DATA_FILE="/data/0_0.tigerbeetle"

# Check if data file exists; if not, format it
if [ ! -f "$TB_DATA_FILE" ]; then
    echo "[INIT] Data file not found. Formatting TigerBeetle Cluster 0..."
    
    # FORMAT: Prepare the data file for a single-node development cluster
    /tigerbeetle format --cluster=0 --replica=0 --replica-count=1 "$TB_DATA_FILE"
    
    echo "[INIT] TigerBeetle formatted successfully."
else
    echo "[INIT] Data file found. Skipping format."
fi

echo "[INIT] Starting TigerBeetle on 0.0.0.0:3000..."

# 'exec' is vital: It replaces the shell process with TigerBeetle (PID 1).
# This allows TigerBeetle to receive shutdown signals (SIGTERM) correctly,
# preventing data corruption when you run 'docker-compose down'.
exec /tigerbeetle start --development --addresses=0.0.0.0:3000 "$TB_DATA_FILE"