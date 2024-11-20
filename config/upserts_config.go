package config

// Config is the configuration for additional upsert modifications.
type UpsertConfig struct {
	// RestrictedMarkets removes the defined markets from the final set of market upserts.
	// This ensures that a chain's marketmap does not receive updates for the markets defined here.
	RestrictedMarkets []string `json:"restricted_markets"`
}

func DefaultUpsertConfig() UpsertConfig {
	return UpsertConfig{
		RestrictedMarkets: []string{},
	}
}

func (c *UpsertConfig) Validate() error {
	return nil
}
