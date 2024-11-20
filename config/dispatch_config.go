package config

import (
	"fmt"
)

// DispatchConfig represents the dispatcher's config data-structure.
type DispatchConfig struct {
	// TxConfig is the configuration that the market-update provider expects.
	TxConfig TransactionConfig `json:"tx"`

	// SigningConfig is the config for transaction signing.
	SigningConfig SigningConfig `json:"signing"`

	// SubmitterConfig is the configuration that the transaction submitter expects.
	SubmitterConfig SubmitterConfig `json:"submitter"`
}

func DefaultDispatchConfig() DispatchConfig {
	return DispatchConfig{
		TxConfig:        DefaultTxConfig(),
		SigningConfig:   DefaultSigningConfig(),
		SubmitterConfig: DefaultSubmitterConfig(),
	}
}

func (c *DispatchConfig) Validate() error {
	// this is the config for submitting tx to a chain, so we do not need it for
	// dry run message generation
	if err := c.TxConfig.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid market update config: %w", err)
	}

	// submitter config has no context of chains or endpoints, so it can be overwritten
	if err := c.SubmitterConfig.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid submitter config: %w", err)
	}

	if err := c.SigningConfig.Validate(); err != nil {
		return fmt.Errorf("invalid signing config: %w", err)
	}

	return nil
}
