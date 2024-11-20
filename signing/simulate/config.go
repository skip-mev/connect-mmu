package simulate

import "errors"

type SigningAgentConfig struct {
	Address string `json:"address"`
}

func (c *SigningAgentConfig) Validate() error {
	if c.Address == "" {
		return errors.New("address is required")
	}
	return nil
}
