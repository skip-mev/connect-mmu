# Transformer

A transformer performs arbitrary data transformations on
a set of `ProviderMarket` objects.

```go
type Transformer interface {
    Transform(
        ctx context.Context,
        markets []db.ProviderMarket,
    ) ([]db.ProviderMarket, error)
}
```
