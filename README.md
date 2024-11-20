# Connect Market Map Updater

The Connect Market Map Updater indexes market data from exchanges (or market providers) into the format of Connect's Market Map module. It converts that market map into transactions that can update the market map on chain.

The Updater is broken into multiple stages for easier debugging.

## Configs

Some configs are provided for you, but you need a coinmarketcap api key to index asset data from coinmarketcap's api. Once you have a key, add it to the JSON config under "index.coinmarketcap.api_key". Alternative, you can set an environment variable `CMC_API_KEY` to your API key.

The available configs are

```
./local/config-dydx-localnet.json
./local/config-dydx-devnet.json
./local/config-dydx-testnet.json
./local/config-dydx-mainnet.json
```

The localnet config is meant for a locally running testnet. You can configure the chain details under the "chain" key in the config, such as api endpoints.

## Running the workflow

The workflow has five stages. Each stage outputs files which can be inspected to debug issues. They are located in the local `./tmp` directory by default.

Defaults can be customized by flags. Pass `--help` to a command to get details about how to change these defaults.

### Testnet

```bash
export CMC_API_KEY='...'

go run ./cmd/mmu index --config ./local/config-dydx-testnet.json
go run ./cmd/mmu generate --config ./local/config-dydx-testnet.json
go run ./cmd/mmu override --config ./local/config-dydx-testnet.json
go run ./cmd/mmu upserts --config ./local/config-dydx-testnet.json --warn-on-invalid-market-map # market map on chain is invalid on testnet
go run ./cmd/mmu dispatch --config ./local/config-dydx-testnet.json --simulate
```

### Mainnet

```bash
export CMC_API_KEY='...'

go run ./cmd/mmu index --config ./local/config-dydx-mainnet.json
go run ./cmd/mmu generate --config ./local/config-dydx-mainnet.json
go run ./cmd/mmu override --config ./local/config-dydx-mainnet.json
go run ./cmd/mmu upserts --config ./local/config-dydx-mainnet.json
go run ./cmd/mmu dispatch --config ./local/config-dydx-mainnet.json --simulate
```
