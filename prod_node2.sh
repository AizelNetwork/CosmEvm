#!/bin/bash
set -e

# Ensure these environment variables are defined in your environment:
#   AIZELHOME, CHAINID, BASE_DENOM, KEYRING, KEYALGO, BASEFEE, VAL2_KEY
# For example, export them before running this script.
CHAINID="${CHAIN_ID:-aizel_2015-3333}"
BASE_DENOM="aaizel"
MONIKER="node1"
KEYRING="file"
KEYALGO="eth_secp256k1"
LOGLEVEL="info"
BASEFEE=1000000000
VAL2_KEY="validator2"

# Define directories for node1 and node2
NODE1_HOME="$AIZELHOME/node1"
NODE2_HOME="$AIZELHOME/node2"

echo "Deleting node2 folder if it exists..."
sudo rm -rf "$NODE2_HOME"

echo "Copying node1 folder to node2..."
sudo cp -r "$NODE1_HOME" "$NODE2_HOME"

# Change ownership of the newly created node2 folder so your user can modify its files.
echo "Changing ownership of node2 folder to the current user..."
sudo chown -R "$(id -u):$(id -g)" "$NODE2_HOME"

# Define configuration file paths for node2
APP_TOML="$NODE2_HOME/config/app.toml"
CONFIG_TOML="$NODE2_HOME/config/config.toml"
CLIENT_TOML="$NODE2_HOME/config/client.toml"

echo "Updating app.toml settings for node2..."

# 1. Change the node name from "node1" to "node2"
sed -i.bak 's/name "node1"/name "node2"/g' "$APP_TOML"

# 2. Update the ports in app.toml
sed -i.bak 's/:11317/:11327/g' "$APP_TOML"
sed -i.bak 's/:18545/:18555/g' "$APP_TOML"
sed -i.bak 's/:18546/:18566/g' "$APP_TOML"
sed -i.bak 's/:56657/:56667/g' "$APP_TOML"
sed -i.bak 's/:56658/:56668/g' "$APP_TOML"
sed -i.bak 's/:56656/:56666/g' "$APP_TOML"
sed -i.bak 's/:16065/:16075/g' "$APP_TOML"

# 3. Change the pruning setting from "nothing" to "custom"
sed -i.bak 's/pruning = "nothing"/pruning = "custom"/g' "$APP_TOML"

echo "Updating config.toml and client.toml settings for node2..."

# Update ports in config.toml and client.toml
for file in "$CONFIG_TOML" "$CLIENT_TOML"; do
  sed -i.bak 's/:56656/:56666/g' "$file"
  sed -i.bak 's/:56657/:56667/g' "$file"
  sed -i.bak 's/:56658/:56668/g' "$file"
  sed -i.bak 's/:16060/:16070/g' "$file"
done

echo "Setting persistent_peers in node2's config.toml..."

# Get the boot node's (node1) node ID
BOOT_NODE_ID=$(aizeld tendermint show-node-id --home "$NODE1_HOME")
echo "Node1 ID: ${BOOT_NODE_ID}"

# Set the boot node's IP address.
# Replace this with your actual node1 IP address if it's not localhost.
BOOT_NODE_IP="127.0.0.1"

# Update the persistent_peers entry in node2's config.toml.
sed -i.bak "s/^persistent_peers = .*/persistent_peers = \"${BOOT_NODE_ID}@${BOOT_NODE_IP}:56656\"/g" "$CONFIG_TOML"

echo "Removing existing gentx files from node2 config/gentx..."
rm -f "$NODE2_HOME/config/gentx/"*.json

echo "Signing genesis transaction for validator2 on node2..."

# Sign a gentx for validator2 on node2.
aizeld gentx "$VAL2_KEY" 1000000000000000000000$BASE_DENOM \
  --gas-prices ${BASEFEE}$BASE_DENOM \
  --keyring-backend "$KEYRING" \
  --chain-id "$CHAINID" \
  --home "$NODE2_HOME"

echo "Copying gentx files from node2 to node1..."

# Copy the gentx files from node2 into node1's gentx folder.
sudo cp -r "$NODE2_HOME/config/gentx/"* "$NODE1_HOME/config/gentx/"

echo "node2 configuration complete."
