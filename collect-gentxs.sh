#!/bin/bash
set -e

CHAINID="${CHAIN_ID:-aizel_2015-3333}"
BASE_DENOM="aaizel"
MONIKER="node1"
# Remember to change to other types of keyring like 'file' in-case exposing to outside world,
# otherwise your balance will be wiped quickly.
# The keyring test does not require private key to steal tokens from you.
KEYRING="file"
KEYALGO="eth_secp256k1"
LOGLEVEL="info"
# Set dedicated home directory for the aizeld instance
HOMEDIR="$AIZELHOME/node1"
# to trace evm
#TRACE="--trace"
TRACE=""

# feemarket params basefee
BASEFEE=1000000000

# Path variables
CONFIG=$HOMEDIR/config/config.toml
APP_TOML=$HOMEDIR/config/app.toml
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json

# Collect genesis tx
aizeld collect-gentxs --home "$HOMEDIR"

# Run this to ensure everything worked and that the genesis file is setup correctly
aizeld validate-genesis --home "$HOMEDIR"

# --- Additional Logic: Copy genesis file from node1 to node2 ---
echo "Copying genesis file from node1 to node2..."

# Remove node2's current genesis file
sudo rm -f "$AIZELHOME/node2/config/genesis.json"

# Copy the genesis file from node1 to node2
sudo cp "$AIZELHOME/node1/config/genesis.json" "$AIZELHOME/node2/config/"

echo "Genesis file successfully copied to node2."
