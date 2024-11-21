#!/bin/bash

set -euxo pipefail

# setup some vars to refer to the root of the directory. i.e. sauron
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DYDXDIR=v4-chain
if [ -z "${DYDXREF:-}" ]; then
  DYDXREF="main"
fi

wait_for_server() {
    echo "waiting for dydx node to become available..."
    local max_attempts=60
    local attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if curl -s -o /dev/null http://localhost:1317; then
            echo "dydx node is ready"
            return 0
        fi
        attempt=$((attempt+1))
        echo "attempt $attempt/$max_attempts: node not ready, waiting..."
        sleep 5
    done
    echo "dydx node did not become available in time."
    return 1
}

cd_to_root() {
    cd "$PROJECT_ROOT"
}

# make sure we're in root before we gogo
cd_to_root

if [ ! -d "$DYDXDIR" ]; then
  echo "$DYDXDIR does not exist. Cloning dydx repo."
  git clone https://github.com/dydxprotocol/v4-chain.git
fi

cd v4-chain

git checkout -- ./protocol/testing
git checkout $DYDXREF
echo $'520a\ndasel put -t json -f "$GENESIS" \'.app_state.marketmap.params.market_authorities\' -v "[\\\"dydx199tqg4wdlnu4qjlxchpd7seg454937hjrknju4\\"]"\n.\nw' | ed -s protocol/testing/genesis.sh

# insert the current mainnet state as the marketmap state
echo $'1273a\napk add --no-cache curl\n.\nw' | ed -s protocol/testing/genesis.sh
echo $'1274a\ncurl https://dydx-api.polkachu.com/slinky/marketmap/v1/marketmap | jq \'.market_map.markets |= map_values(if .ticker then .ticker.decimals = (.ticker.decimals | tonumber) | .ticker.min_provider_count = (.ticker.min_provider_count | tonumber) else . end) | .market_map\' > temp_market_map.json\n.\nw' | ed -s protocol/testing/genesis.sh
echo $'1275a\njq --argfile new_data temp_market_map.json ".app_state.marketmap.market_map = \$new_data" "$GENESIS" > temp.json && mv temp.json "$GENESIS"\n.\nw' | ed -s protocol/testing/genesis.sh
echo $'1276a\nrm temp_market_map.json\n.\nw' | ed -s protocol/testing/genesis.sh

cd protocol

make localnet-startd

wait_for_server
