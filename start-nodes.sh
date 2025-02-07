#!/bin/bash
CHAINID="${CHAIN_ID:-aizel_2015-3333}"
BASE_DENOM="aaizel"
TRACE=""
LOGLEVEL="info"
# Start the node1
aizeld start \
	--metrics "$TRACE" \
	--log_level $LOGLEVEL \
	--minimum-gas-prices=0.0001$BASE_DENOM \
	--json-rpc.api eth,txpool,personal,net,debug,web3 \
	--home "$AIZELHOME/node1" \
	--chain-id "$CHAINID" \
	 > $AIZELHOME/node1/node1.log 2>&1 &

sleep 2

# Start the node2
aizeld start \
	--metrics "$TRACE" \
	--log_level $LOGLEVEL \
	--minimum-gas-prices=0.0001$BASE_DENOM \
	--json-rpc.api eth,txpool,personal,net,debug,web3 \
	--home "$AIZELHOME/node2" \
	--chain-id "$CHAINID" \
	 > $AIZELHOME/node2/node2.log 2>&1 &

# tmp
# aizeld query ibc client params --chain-id aizel_9002-20151225