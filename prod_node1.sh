#!/bin/bash

CHAINID="${CHAIN_ID:-aizel_2015-3333}"
BASE_DENOM="aaizel"
MONIKER="node1"
# Remember to change to other types of keyring like 'file' in-case exposing to outside world,
# otherwise your balance will be wiped quickly
# The keyring test does not require private key to steal tokens from you
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

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

# used to exit on first error (any non-zero exit code)
set -e

# Parse input flags
install=true
overwrite=""

while [[ $# -gt 0 ]]; do
	key="$1"
	case $key in
	-y)
		echo "Flag -y passed -> Overwriting the previous chain data."
		overwrite="y"
		shift # Move past the flag
		;;
	-n)
		echo "Flag -n passed -> Not overwriting the previous chain data."
		overwrite="n"
		shift # Move past the argument
		;;
	--no-install)
		echo "Flag --no-install passed -> Skipping installation of the aizeld binary."
		install=false
		shift # Move past the flag
		;;
	*)
		echo "Unknown flag passed: $key -> Exiting script!"
		exit 1
		;;
	esac
done

if [[ $install == true ]]; then
	# (Re-)install daemon
	make install
fi

# User prompt if neither -y nor -n was passed as a flag
# and an existing local node configuration is found.
if [[ $overwrite = "" ]]; then
	if [ -d "$HOMEDIR" ]; then
		printf "\nAn existing folder at '%s' was found. You can choose to delete this folder and start a new local node with new keys from genesis. When declined, the existing local node is started. \n" "$HOMEDIR"
		echo "Overwrite the existing configuration and start a new local node? [y/n]"
		read -r overwrite
	else
		overwrite="y"
	fi
fi

# Setup local node if overwrite is set to Yes, otherwise skip setup
if [[ $overwrite == "y" || $overwrite == "Y" ]]; then
	# Remove the previous folder
	rm -rf "$HOMEDIR"

	# Set client config
	aizeld config set client chain-id "$CHAINID" --home "$HOMEDIR"
	aizeld config set client keyring-backend "$KEYRING" --home "$HOMEDIR"

	# validator1 address 0x | aizel1yxkkj0ujpt95mj7zfy737ua8tqtgxhqft5ze93
	VAL1_KEY="validator1"
	VAL1_MNEMONIC="razor system inherit front boss cigar youth museum vocal enhance fat dolphin gas joy gift peace ramp doctor cargo equip chalk joke arrow fresh"
	VAL1_KEY_PASS="20151225"
	VAL2_KEY="validator2"
	VAL2_MNEMONIC="surround soft tragic ensure accuse tooth soul attack ahead cheese few taxi hope priority globe crater enable still silly glad hire correct brush defense"
	VAL2_KEY_PASS="20151225"
	# dev0 address 0xaaafB3972B05630fCceE866eC69CdADd9baC2771 | aizel142hm89etq43sln8wsehvd8x6mkd6cfm35279kg
	USER1_KEY_ADDRESS="aizel142hm89etq43sln8wsehvd8x6mkd6cfm35279kg"
	
	# # dev1 address 0xF34B930cF45ED5C1B095A9ED7f9F9b63676C31a5 | aizel17d9exr85tm2urvy448khl8umvdnkcvd9hc7dpl
	USER2_KEY_ADDRESS="aizel17d9exr85tm2urvy448khl8umvdnkcvd9hc7dpl"
	
	# # dev2 address 0x6Be02d1d3665660d22FF9624b7BE0551ee1Ac91b | aizel1d0sz68fkv4nq6ghljcjt00s928hp4jgm2egveq
	USER3_KEY_ADDRESS="aizel1d0sz68fkv4nq6ghljcjt00s928hp4jgm2egveq"
	
	# # dev3 address 0xFcC0E188e0214B818AFD4a8aA71EaEF7dcaf8ffd | aizel1lnqwrz8qy99crzhaf292w84w7lw2lrlalxhdmy
	USER4_KEY_ADDRESS="aizel1lnqwrz8qy99crzhaf292w84w7lw2lrlalxhdmy"
	
	# Import keys from mnemonics
	aizeld keys add "$VAL1_KEY" \
  		--recover \
  		--keyring-backend file \
  		--algo "$KEYALGO" \
  		--home "$HOMEDIR"<<EOF
$VAL1_MNEMONIC
$VAL1_KEY_PASS
$VAL1_KEY_PASS
EOF
	aizeld keys add "$VAL2_KEY" \
  		--recover \
  		--keyring-backend file \
  		--algo "$KEYALGO" \
  		--home "$HOMEDIR"<<EOF
$VAL2_MNEMONIC
$VAL2_KEY_PASS
$VAL2_KEY_PASS
EOF

	# Set moniker and chain-id for Aizel (Moniker can be anything, chain-id must be an integer)
	aizeld init $MONIKER -o --chain-id "$CHAINID" --home "$HOMEDIR"

	# Change parameter token denominations to $BASE_DENOM
	jq --arg base_denom "$BASE_DENOM" '.app_state["staking"]["params"]["bond_denom"]=$base_denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq --arg base_denom "$BASE_DENOM" '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]=$base_denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	# When upgrade to cosmos-sdk v0.47, use gov.params to edit the deposit params
	jq --arg base_denom "$BASE_DENOM" '.app_state["gov"]["params"]["min_deposit"][0]["denom"]=$base_denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq --arg base_denom "$BASE_DENOM" '.app_state["inflation"]["params"]["mint_denom"]=$base_denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Set gas limit in genesis
	jq '.consensus_params["block"]["max_gas"]="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	# Set base fee in genesis
	jq '.app_state["feemarket"]["params"]["base_fee"]="'${BASEFEE}'"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

	if [[ $1 == "pending" ]]; then
		if [[ "$OSTYPE" == "darwin"* ]]; then
			sed -i '' 's/timeout_propose = "3s"/timeout_propose = "30s"/g' "$CONFIG"
			sed -i '' 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' "$CONFIG"
			sed -i '' 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' "$CONFIG"
			sed -i '' 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' "$CONFIG"
			sed -i '' 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' "$CONFIG"
			sed -i '' 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' "$CONFIG"
			sed -i '' 's/timeout_commit = "5s"/timeout_commit = "150s"/g' "$CONFIG"
			sed -i '' 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' "$CONFIG"
		else
			sed -i 's/timeout_propose = "3s"/timeout_propose = "30s"/g' "$CONFIG"
			sed -i 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' "$CONFIG"
			sed -i 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' "$CONFIG"
			sed -i 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' "$CONFIG"
			sed -i 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' "$CONFIG"
			sed -i 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' "$CONFIG"
			sed -i 's/timeout_commit = "5s"/timeout_commit = "150s"/g' "$CONFIG"
			sed -i 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' "$CONFIG"
		fi
	fi

	# enable prometheus metrics and all APIs for dev node
	if [[ "$OSTYPE" == "darwin"* ]]; then
		sed -i '' 's/prometheus = false/prometheus = true/' "$CONFIG"
		sed -i '' 's/prometheus-retention-time = 0/prometheus-retention-time  = 1000000000000/g' "$APP_TOML"
		sed -i '' 's/enabled = false/enabled = true/g' "$APP_TOML"
		sed -i '' 's/enable = false/enable = true/g' "$APP_TOML"
		# Don't enable Rosetta API by default
		grep -q -F '[rosetta]' "$APP_TOML" && sed -i '' '/\[rosetta\]/,/^\[/ s/enable = true/enable = false/' "$APP_TOML"
		# Don't enable memiavl by default
		grep -q -F '[memiavl]' "$APP_TOML" && sed -i '' '/\[memiavl\]/,/^\[/ s/enable = true/enable = false/' "$APP_TOML"
		# Don't enable versionDB by default
		grep -q -F '[versiondb]' "$APP_TOML" && sed -i '' '/\[versiondb\]/,/^\[/ s/enable = true/enable = false/' "$APP_TOML"
	else
		sed -i 's/prometheus = false/prometheus = true/' "$CONFIG"
		sed -i 's/prometheus-retention-time  = "0"/prometheus-retention-time  = "1000000000000"/g' "$APP_TOML"
		sed -i 's/enabled = false/enabled = true/g' "$APP_TOML"
		sed -i 's/enable = false/enable = true/g' "$APP_TOML"
		# Don't enable Rosetta API by default
		grep -q -F '[rosetta]' "$APP_TOML" && sed -i '/\[rosetta\]/,/^\[/ s/enable = true/enable = false/' "$APP_TOML"
		# Don't enable memiavl by default
		grep -q -F '[memiavl]' "$APP_TOML" && sed -i '/\[memiavl\]/,/^\[/ s/enable = true/enable = false/' "$APP_TOML"
		# Don't enable versionDB by default
		grep -q -F '[versiondb]' "$APP_TOML" && sed -i '/\[versiondb\]/,/^\[/ s/enable = true/enable = false/' "$APP_TOML"
	fi

	# Additional modifications as requested:
	if [[ "$OSTYPE" == "darwin"* ]]; then
		# 1 & 2: Replace "localhost" & "127.0.0.1" with "0.0.0.0" in app, client, genesis
		sed -i '' 's/localhost/0.0.0.0/g' "$APP_TOML"
		sed -i '' 's/localhost/0.0.0.0/g' "$HOMEDIR/config/client.toml" 2>/dev/null || true
		sed -i '' 's/localhost/0.0.0.0/g' "$GENESIS"
		sed -i '' 's/create_0.0.0.0/create_localhost/g' "$GENESIS"
		sed -i '' 's/127.0.0.1/0.0.0.0/g' "$APP_TOML"
		sed -i '' 's/127.0.0.1/0.0.0.0/g' "$HOMEDIR/config/client.toml" 2>/dev/null || true
		sed -i '' 's/127.0.0.1/0.0.0.0/g' "$GENESIS"

		# 3: enabled-unsafe-cors = true in app
		sed -i '' 's/^enabled-unsafe-cors = .*/enabled-unsafe-cors = true/' "$APP_TOML"

		# 5: pruning="nothing" in app
		sed -i '' 's/^pruning = .*/pruning = "nothing"/' "$APP_TOML"

		# 6,7,8,9,10,11 (ports in app)
		sed -i '' 's/:1317/:11317/g' "$APP_TOML"
		sed -i '' 's/:9090/:19090/g' "$APP_TOML"
		sed -i '' 's/:8545/:18545/g' "$APP_TOML"
		sed -i '' 's/:8546/:18546/g' "$APP_TOML"
		sed -i '' 's/:6065/:16065/g' "$APP_TOML"
		sed -i '' 's/:26657/:56657/g' "$APP_TOML"

		# 11,12,13,14,15 (ports in config + client)
		sed -i '' 's/:26657/:56657/g' "$CONFIG"
		sed -i '' 's/:26658/:56658/g' "$CONFIG"
		sed -i '' 's/:6060/:16060/g' "$CONFIG"
		sed -i '' 's/:26656/:56656/g' "$CONFIG"
		sed -i '' 's/:26660/:56660/g' "$CONFIG"
		# Also client
		sed -i '' 's/:26657/:56657/g' "$HOMEDIR/config/client.toml" 2>/dev/null || true

		# 4: cors_allowed_origins = ["*"] in config
		sed -i '' 's#^cors_allowed_origins = .*#cors_allowed_origins = ["*"]#' "$CONFIG"
	else
		# 1 & 2: Replace "localhost" & "127.0.0.1" with "0.0.0.0" in app, client, genesis
		sed -i 's/localhost/0.0.0.0/g' "$APP_TOML"
		sed -i 's/localhost/0.0.0.0/g' "$HOMEDIR/config/client.toml" 2>/dev/null || true
		sed -i 's/localhost/0.0.0.0/g' "$GENESIS"
		sed -i 's/create_0.0.0.0/create_localhost/g' "$GENESIS"
		sed -i 's/127.0.0.1/0.0.0.0/g' "$APP_TOML"
		sed -i 's/127.0.0.1/0.0.0.0/g' "$HOMEDIR/config/client.toml" 2>/dev/null || true
		sed -i 's/127.0.0.1/0.0.0.0/g' "$GENESIS"

		# 3: enabled-unsafe-cors = true in app
		sed -i 's/^enabled-unsafe-cors = .*/enabled-unsafe-cors = true/' "$APP_TOML"

		# 5: pruning="nothing" in app
		sed -i 's/^pruning = .*/pruning = "nothing"/' "$APP_TOML"

		# 6,7,8,9,10,11 (ports in app)
		sed -i 's/:1317/:11317/g' "$APP_TOML"
		sed -i 's/:9090/:19090/g' "$APP_TOML"
		sed -i 's/:8545/:18545/g' "$APP_TOML"
		sed -i 's/:8546/:18546/g' "$APP_TOML"
		sed -i 's/:6065/:16065/g' "$APP_TOML"
		sed -i 's/:26657/:56657/g' "$APP_TOML"

		# 11,12,13,14,15 (ports in config + client)
		sed -i 's/:26657/:56657/g' "$CONFIG"
		sed -i 's/:26658/:56658/g' "$CONFIG"
		sed -i 's/:6060/:16060/g' "$CONFIG"
		sed -i 's/:26656/:56656/g' "$CONFIG"
		sed -i 's/:26660/:56660/g' "$CONFIG"
		# Also client
		sed -i 's/:26657/:56657/g' "$HOMEDIR/config/client.toml" 2>/dev/null || true

		# 4: cors_allowed_origins = ["*"] in config
		sed -i 's#^cors_allowed_origins = .*#cors_allowed_origins = ["*"]#' "$CONFIG"
	fi

	# Change proposal periods to pass within a reasonable time for local testing
	sed -i.bak 's/"max_deposit_period": "172800s"/"max_deposit_period": "30s"/g' "$GENESIS"
	sed -i.bak 's/"voting_period": "172800s"/"voting_period": "30s"/g' "$GENESIS"
	sed -i.bak 's/"expedited_voting_period": "86400s"/"expedited_voting_period": "15s"/g' "$GENESIS"

	# set custom pruning settings
	sed -i.bak 's/pruning = "default"/pruning = "custom"/g' "$APP_TOML"
	sed -i.bak 's/pruning-keep-recent = "0"/pruning-keep-recent = "2"/g' "$APP_TOML"
	sed -i.bak 's/pruning-interval = "0"/pruning-interval = "10"/g' "$APP_TOML"

	# Allocate genesis accounts (cosmos formatted addresses)
	aizeld add-genesis-account "$(aizeld keys show "$VAL1_KEY" -a --keyring-backend "$KEYRING" --home "$HOMEDIR")" 100000000000000000000000000$BASE_DENOM --keyring-backend "$KEYRING" --home "$HOMEDIR"
	aizeld add-genesis-account "$(aizeld keys show "$VAL2_KEY" -a --keyring-backend "$KEYRING" --home "$HOMEDIR")" 100000000000000000000000000$BASE_DENOM --keyring-backend "$KEYRING" --home "$HOMEDIR"
	aizeld add-genesis-account "$USER1_KEY_ADDRESS" 1000000000000000000000000000$BASE_DENOM --keyring-backend "$KEYRING" --home "$HOMEDIR"
	aizeld add-genesis-account "$USER2_KEY_ADDRESS" 1000000000000000000000000000$BASE_DENOM --keyring-backend "$KEYRING" --home "$HOMEDIR"
	aizeld add-genesis-account "$USER3_KEY_ADDRESS" 1000000000000000000000000000$BASE_DENOM --keyring-backend "$KEYRING" --home "$HOMEDIR"
	aizeld add-genesis-account "$USER4_KEY_ADDRESS" 1000000000000000000000000000$BASE_DENOM --keyring-backend "$KEYRING" --home "$HOMEDIR"
	aizeld gentx "$VAL1_KEY" 1000000000000000000000$BASE_DENOM --gas-prices ${BASEFEE}$BASE_DENOM --keyring-backend "$KEYRING" --chain-id "$CHAINID" --home "$HOMEDIR"

	if [[ $1 == "pending" ]]; then
		echo "pending mode is on, please wait for the first block committed."
	fi
fi

# Start the node
# aizeld start \
# 	--metrics "$TRACE" \
# 	--log_level $LOGLEVEL \
# 	--minimum-gas-prices=0.0001$BASE_DENOM \
# 	--json-rpc.api eth,txpool,personal,net,debug,web3 \
# 	--home "$HOMEDIR" \
# 	--chain-id "$CHAINID"