#!/bin/bash
CHAINID="${CHAIN_ID:-aizel_2015-3333}"
BASE_DENOM="aaizel"
VAL1_KEY="validator1"
# Submit Proposal
aizeld tx gov submit-proposal "$AIZELHOME/evm-v9-upgrade-proposal.json" \
  --fees 4430808153$BASE_DENOM \
  --home "$AIZELHOME/node1" \
  --chain-id "$CHAINID" \
  --from "$VAL1_KEY" \
  --yes \
  > "$AIZELHOME/submit-proposal.log"



  # aizeld tx upgrade software-upgrade v20 \
  # --fees 4430808153aaizel \
  # --title="Test Proposal"  \
  # --summary="testing" \
  # --deposit="1000000aaizel"  \
  # --upgrade-height 1000000 \
  # --upgrade-info '{ "binaries": {"linux/arm64":"https://github.com/AizelNetwork/CosmEvm/releases/download/v0.0.2/aizeld?checksum=sha256:f4953ed8705bfff10386f68f3ed88a1614ea21a4166f00c4b7293f9b421170a9"} }'  \
  # --from "validator1" \
  # --home "$AIZELHOME/node1" \
  # --chain-id "aizel_2015-3333"
