#!/bin/bash
CHAINID="${CHAIN_ID:-aizel_2015-3333}"
BASE_DENOM="aaizel"
VAL1_KEY="validator1"
# Submit Proposal
aizeld query gov proposal 5 \
  --home "$AIZELHOME/node1" \
  --chain-id "$CHAINID" \
  > "$AIZELHOME/query-proposal.log"
