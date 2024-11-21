package local

import (
	"fmt"
	"os"
)

type SigningAgentConfig struct {
	PrivateKeyFile string `json:"private_key_file"`
}

func (c *SigningAgentConfig) Validate() error {
	if _, err := os.Stat(c.PrivateKeyFile); os.IsNotExist(err) {
		return fmt.Errorf("private key file (%s) does not exist", c.PrivateKeyFile)
	}

	return nil
}
