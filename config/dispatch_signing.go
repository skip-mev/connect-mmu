package config

type SigningConfig struct {
	Type   string `json:"type"`
	Config any    `json:"config"`
}

func (s *SigningConfig) Validate() error {
	return nil
}

func DefaultSigningConfig() SigningConfig {
	return SigningConfig{
		Type:   "",
		Config: nil,
	}
}
