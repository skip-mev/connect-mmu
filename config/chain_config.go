package config

import (
	"fmt"
)

// ChainConfig is a configuration for a chain.
type ChainConfig struct {
	// RPCAddress is the address of the chain's RPC server
	RPCAddress string `json:"rpc_address"`

	// GRPCAddress is the address of the chain's GRPC server
	GRPCAddress string `json:"grpc_address"`

	// RESTAddress is the address of the chain's REST server
	RESTAddress string `json:"rest_address"`

	// ChainID is the chain id of the chain.
	ChainID string `json:"chain_id"`

	// DYDX is a bool that indicates if the chain is a dydx chain
	DYDX bool `json:"dydx"`

	// Version is the version of Connect (slinky or connect) this chain uses.
	Version Version `json:"version"`

	// Prefix is the address prefix the chain uses (ex: cosmos).
	Prefix string `json:"prefix"`
}

func DefaultChainConfig() ChainConfig {
	return ChainConfig{
		RPCAddress:  "http://localhost:26657",
		GRPCAddress: "localhost:9090",
		RESTAddress: "http://localhost:1317",
		DYDX:        false,
		Version:     VersionSlinky,
		Prefix:      "cosmos",
	}
}

func (c *ChainConfig) Validate() error {
	if c.GRPCAddress == "" || c.RESTAddress == "" || c.RPCAddress == "" {
		return NewErrInvalidChainConfig(fmt.Errorf("invalid chain config: rest, rpc or grpc address is empty: %s, %s", c.GRPCAddress, c.RESTAddress))
	}

	if !IsValidVersion(c.Version) {
		return fmt.Errorf("version must be one of (%s, %s)", VersionConnect, VersionSlinky)
	}

	if c.ChainID == "" {
		return NewErrInvalidChainConfig(fmt.Errorf("invalid chain config: chain_id is empty"))
	}

	if c.Prefix == "" {
		return fmt.Errorf("invalid chain config: prefix is empty, %s", c.Prefix)
	}

	return nil
}

// ErrInvalidChainConfig is an error that occurs when the chain config is invalid.
type ErrInvalidChainConfig struct {
	err error
}

// Error implements the error interface.
func (e ErrInvalidChainConfig) Error() string {
	return fmt.Errorf("invalid chain config: %w", e.err).Error()
}

// NewErrInvalidChainConfig creates a new ErrInvalidChainConfig.
func NewErrInvalidChainConfig(err error) ErrInvalidChainConfig {
	return ErrInvalidChainConfig{err: err}
}
