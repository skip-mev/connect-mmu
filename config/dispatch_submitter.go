package config

import (
	"fmt"
	"time"
)

const (
	DefaultPollingFrequency = time.Second * 10
	DefaultPollingDuration  = time.Minute * 5
)

// SubmitterConfig is the configuration for a transaction submitter.
type SubmitterConfig struct {
	// PollingFrequency is the frequency at which the submitter polls for transaction results.
	PollingFrequency time.Duration `json:"polling_frequency"`

	// PollingDuration is the total duration the submitter polls for transaction results.
	PollingDuration time.Duration `json:"polling_duration"`
}

func DefaultSubmitterConfig() SubmitterConfig {
	return SubmitterConfig{
		PollingFrequency: DefaultPollingFrequency,
		PollingDuration:  DefaultPollingDuration,
	}
}

func (c *SubmitterConfig) ValidateBasic() error {
	if c.PollingFrequency == 0 {
		return fmt.Errorf("polling_frequency must be greater than zero")
	}

	if c.PollingDuration == 0 {
		return fmt.Errorf("polling_duration must be greater than zero")
	}

	return nil
}
