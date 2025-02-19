#!/bin/bash
CHAINID="${CHAIN_ID:-aizel_2015-3333}"
BASE_DENOM="aaizel"
VAL2_KEY="validator2"
# Submit Proposal
aizeld tx gov vote 5 yes \
  --fees 4430808153$BASE_DENOM \
  --home "$AIZELHOME/node2" \
  --chain-id "$CHAINID" \
  --from "$VAL2_KEY" \
  --yes \
  > "$AIZELHOME/vote-proposal.log"


  # aizeld tx gov vote 1 yes --from validator2 --home "$AIZELHOME/node2" --fees 4430808153aaizel
