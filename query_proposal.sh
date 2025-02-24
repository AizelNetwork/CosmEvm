#!/bin/bash
CHAINID="${CHAIN_ID:-aizel_2015-3333}"
VAL1_KEY="validator1"
# Submit Proposal
aizeld query gov proposal 1 \
  --home "$AIZELHOME/node1" \
  --chain-id "$CHAINID" \
  --node "tcp://localhost:26657" \
  > "$AIZELHOME/query-proposal.log"

# aizeld query gov proposals \
#   --home "$AIZELHOME/node1" \
#   --node "tcp://localhost:26657"
