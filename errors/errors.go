package errors

import (
	"fmt"
)

// MarketNotFoundError is an error thrown from market-map validator + delta-checking that indicates
// a market that should have existed in a market-map, is not actually there.
type MarketNotFoundError struct {
	msg string
}

func (e MarketNotFoundError) Error() string {
	return fmt.Sprintf("market not found: %s", e.msg)
}

// NewMarketNotFoundError creates a new MarketNotFoundError.
func NewMarketNotFoundError(msg string) MarketNotFoundError {
	return MarketNotFoundError{msg: msg}
}

// InvalidMarketMapError is an error thrown when a market-map is invalid.
type InvalidMarketMapError struct {
	err error
}

func (e InvalidMarketMapError) Error() string {
	return fmt.Sprintf("invalid market-map: %s", e.err)
}

// NewInvalidMarketMapError creates a new InvalidMarketMapError.
func NewInvalidMarketMapError(err error) InvalidMarketMapError {
	return InvalidMarketMapError{err: err}
}

// InvalidGeneratedMarketMapError is an error thrown when an expected market-map is invalid.
type InvalidGeneratedMarketMapError struct {
	err error
}

func (e InvalidGeneratedMarketMapError) Error() string {
	return fmt.Sprintf("invalid expected market-map: %s", e.err)
}

// NewInvalidGeneratedMarketMapError creates a new InvalidGeneratedMarketMapError.
func NewInvalidGeneratedMarketMapError(err error) InvalidGeneratedMarketMapError {
	return InvalidGeneratedMarketMapError{err: err}
}

// InvalidActualMarketMapError is an error thrown when an actual market-map is invalid.
type InvalidActualMarketMapError struct {
	err error
}

func (e InvalidActualMarketMapError) Error() string {
	return fmt.Sprintf("invalid actual market-map: %s", e.err)
}

// NewInvalidActualMarketMapError creates a new InvalidActualMarketMapError.
func NewInvalidActualMarketMapError(err error) InvalidActualMarketMapError {
	return InvalidActualMarketMapError{err: err}
}
