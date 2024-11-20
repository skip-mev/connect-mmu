package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Index    *MarketConfig   `json:"index,omitempty"`
	Generate *GenerateConfig `json:"generate,omitempty"`
	Upsert   *UpsertConfig   `json:"upsert,omitempty"`
	Dispatch *DispatchConfig `json:"dispatch,omitempty"`
	Chain    *ChainConfig    `json:"chain,omitempty"`
}

func (c *Config) Validate() error {
	if c.Index != nil {
		if err := c.Index.Validate(); err != nil {
			return err
		}
	}

	if c.Generate != nil {
		if err := c.Generate.Validate(); err != nil {
			return err
		}
	}

	if c.Upsert != nil {
		if err := c.Upsert.Validate(); err != nil {
			return err
		}
	}

	if c.Dispatch != nil {
		if err := c.Dispatch.Validate(); err != nil {
			return err
		}
	}

	if c.Chain != nil {
		if err := c.Chain.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func DefaultConfig() Config {
	return Config{
		Index:    &[]MarketConfig{DefaultMarketConfig()}[0],
		Generate: &[]GenerateConfig{DefaultGenerateConfig()}[0],
		Upsert:   &[]UpsertConfig{DefaultUpsertConfig()}[0],
		Dispatch: &[]DispatchConfig{DefaultDispatchConfig()}[0],
		Chain:    &[]ChainConfig{DefaultChainConfig()}[0],
	}
}

func WriteConfig(cfg Config, path string) error {
	bz, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, bz, 0o600)
}

func ReadConfig(path string) (Config, error) {
	var cfg Config
	bz, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err = json.Unmarshal(bz, &cfg); err != nil {
		return cfg, err
	}

	return cfg, cfg.Validate()
}
