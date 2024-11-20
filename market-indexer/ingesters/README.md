# Market Data Ingesters

Market data ingesters are packages that implement the `Ingester` interface:

```go
// Ingester is a general interface for a module that can ingest and parse 
// market  data for  a provider (kraken, uniswap, etc.) to determine what 
// markets that provider supports.
type Ingester interface {
    // GetProviderMarkets returns a list of CreateProviderMarketParams for 
    // the given Ingester. 
    GetProviderMarkets(ctx context.Context) (
        []db.CreateProviderMarketParams, 
        error, 
    )
}
```

## Supported Ingesters

The following ingesters are supported:

- [binance](./binance/README.md)
- [bitfinex](./bitfinex/README.md)
- [crypto.com](./crypto.com/README.md)
- [kraken](./kraken/README.md)
