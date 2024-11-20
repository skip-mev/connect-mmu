package signing

import (
	"errors"
	"sync"

	"github.com/skip-mev/connect-mmu/config"
)

type Factory func(cfg any, chainCfg config.ChainConfig) (SigningAgent, error)

// Registry manages Signer factories
type Registry struct {
	mu      sync.RWMutex
	signers map[string]Factory
}

// NewRegistry creates a new Registry instance
func NewRegistry() *Registry {
	return &Registry{
		signers: make(map[string]Factory),
	}
}

func (r *Registry) RegisterSigner(typ string, factory Factory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.signers[typ]; exists {
		return errors.New("signer type already registered: " + typ)
	}

	r.signers[typ] = factory
	return nil
}

func (r *Registry) CreateSigner(cfg config.SigningConfig, chainCfg config.ChainConfig) (SigningAgent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.signers[cfg.Type]
	if !exists {
		return nil, errors.New("unknown signer type: " + cfg.Type)
	}

	return factory(cfg.Config, chainCfg)
}
