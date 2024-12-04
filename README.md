# Connect Market Map Updater

The Connect Market Map Updater indexes market data from exchanges (or market providers) into the format of Connect's Market Map module. It converts that market map into transactions that can update the market map on-chain.

The Updater is split into multiple stages for easier debugging.

## Configurations

Some configurations are provided for you, but you need a CoinMarketCap API key to index asset data from CoinMarketCap's API. Once you have a key, add it to the JSON config under `index.coinmarketcap.api_key`. Alternatively, you can set an environment variable `CMC_API_KEY` to your API key.

The available configurations are:

```
./local/config-dydx-localnet.json
./local/config-dydx-devnet.json
./local/config-dydx-testnet.json
./local/config-dydx-mainnet.json
```

The `localnet` configuration is meant for a locally running testnet. You can configure the chain details under the `chain` key in the config, such as API endpoints.

## Running the Workflow

- **Extra Flags**: Additional command flags are available in `flags.go`, or use `--help` from the CLI to see all options.
- **Output Directory**: By default, all generated files are saved in the `./tmp` directory.


### Testnet Example

```bash
export CMC_API_KEY='...'

go run ./cmd/mmu index --config ./local/config-dydx-testnet.json
go run ./cmd/mmu generate --config ./local/config-dydx-testnet.json
go run ./cmd/mmu override --config ./local/config-dydx-testnet.json
go run ./cmd/mmu upserts --config ./local/config-dydx-testnet.json --warn-on-invalid-market-map # Market map on chain is invalid on testnet
go run ./cmd/mmu dispatch --config ./local/config-dydx-testnet.json --simulate
```

### Mainnet Example

```bash
export CMC_API_KEY='...'

go run ./cmd/mmu index --config ./local/config-dydx-mainnet.json
go run ./cmd/mmu generate --config ./local/config-dydx-mainnet.json
go run ./cmd/mmu override --config ./local/config-dydx-mainnet.json
go run ./cmd/mmu upserts --config ./local/config-dydx-mainnet.json
go run ./cmd/mmu dispatch --config ./local/config-dydx-mainnet.json --simulate
```


## Recommended Workflow

1. **Run `index` and `generate` commands.**
2. **Validate the generated data.**
3. **Apply override rules if necessary.**
4. **Run `upserts` and inspect the output.**
5. **Use `dispatch` with `--simulate` to test the transactions.**
6. **If simulations are successful, run `dispatch` to submit transactions to the blockchain.**

---

## Index

```bash
go run ./cmd/mmu index --config ./local/config-dydx-mainnet.json
```

The `index` job collects asset data from CoinMarketCap and market data from supported providers like Coinbase, Binance, and Uniswap. Its goal is to compile a list of all available markets from these providers.

- **Providers Configuration**: Providers are specified under the `index.ingesters` key in the provider configuration file (e.g., `ingesters`, `coinmarketcap`).
- **API Keys**: Ensure you add your CoinMarketCap API key in the configuration file.

---

## Generate

```bash
go run ./cmd/mmu generate --config ./local/config-dydx-mainnet.json
```

The `generate` job converts provider data into a market map—a collection of base/quote asset pairs (markets). Each market includes metadata (like reference prices) and a list of providers offering prices for that market, each with configuration details. The output is saved as `generated-market-map`.

- **Note**: `generated-market-map-removals` is an additional artifact from the indexing job that contains markets filtered out due to not meeting certain criteria. This is useful for debugging and understanding why some markets were not included.

---

## Validate

Validates configurations and generated market maps. This helps ensure configurations are correct and identifies any transient failures (e.g., API downtime).

```bash
# Clone the Connect repository and checkout the branch or tag you want to install
# Run make install-sentry and add the go binard directory to your shell's PATH
# Run `make install` in the Connect repository and link to Go environment
# sentry in validator/cmd/sentry
# ln -s $(go env GOBIN)/slinky $(go env GOBIN)/connect
go run ./cmd/mmu validate --market-map ./tmp/generated-market-map.json --oracle-config ./local/fixtures/e2e/oracle.json --start-delay 10s --duration 1m
```

---

## Override

```bash
go run ./cmd/mmu override --config ./local/config-dydx-mainnet.json
```

The `override` job compares the generated market map with the current on-chain market map and modifies it based on specific rules. This ensures that certain markets remain unchanged or are updated according to the overrides.

**Optional flags to override default behavior:**

- `--update-enabled`: Updates the on-chain values of enabled markets.
- `--existing-only`: Removes new markets from the market map to prevent adding them.
- `--overwrite-providers`: Maintains existing providers (e.g., Uniswap) without altering their on-chain configurations. Use this to avoid modifying properly configured providers.

---

## Upserts

```bash
go run ./cmd/mmu upserts --config ./local/config-dydx-mainnet.json
```

The `upserts` job examines the market map to identify changed markets and outputs them to a file. These markets are prepared for inclusion in a transaction to update the market map on-chain. This transaction will be submitted as part of the `dispatch` job.

---

## Dispatch

```bash
go run ./cmd/mmu dispatch --config ./local/config-dydx-mainnet.json
```

The `dispatch` job prepares and submits transactions to the blockchain, splitting updates into multiple transactions if necessary due to size constraints.

- **Simulation Recommended**: It's advisable to simulate the transaction before actual submission.
- **Signing Transactions**: Transactions can be signed with local keys saved to disk, but it's recommended to use your own robust signing service.

**Flags:**

- `--simulate`: Simulates the transaction without submitting it. Uses the address configured in `dispatch.signing`.
- `--simulate-address <address>`: Uses a specified address for simulation.

---

## Commands

```bash
go run ./cmd/mmu index --config ./local/config-dydx-mainnet.json
go run ./cmd/mmu generate --config ./local/config-dydx-mainnet.json
go run ./cmd/mmu override --config ./local/config-dydx-mainnet.json
go run ./cmd/mmu upserts --config ./local/config-dydx-mainnet.json
go run ./cmd/mmu dispatch --config ./local/config-dydx-mainnet.json
```
