package generator

import (
	"fmt"
)

// SignerError is an error that occurs when the signer client receives a failed request.
type SignerError struct {
	err error
}

// NewSignerError creates a new signer error.
func NewSignerError(err error) SignerError {
	return SignerError{
		err: err,
	}
}

// Error returns the error message.
func (s SignerError) Error() string {
	return fmt.Sprintf("error from signer client: %s", s.err)
}

// InvalidSignerPubkeyError is an error that occurs when the signer client returns an invalid pubkey.
type InvalidSignerPubkeyError struct {
	err error
}

// NewInvalidSignerPubkeyError creates a new invalid signer pubkey error.
func NewInvalidSignerPubkeyError(err error) InvalidSignerPubkeyError {
	return InvalidSignerPubkeyError{
		err: err,
	}
}

// Error returns the error message.
func (i InvalidSignerPubkeyError) Error() string {
	return fmt.Sprintf("invalid pubkey from signer client: %s", i.err)
}

// AuthClientError is an error that occurs when the auth client receives a failed request.
type AuthClientError struct {
	err error
}

// NewAuthClientError creates a new auth client error.
func NewAuthClientError(err error) AuthClientError {
	return AuthClientError{
		err: err,
	}
}

// Error returns the error message.
func (a AuthClientError) Error() string {
	return fmt.Errorf("error from auth client: %w", a.err).Error()
}

// TxGenerationError is an error that occurs when a transaction fails to be generated.
type TxGenerationError struct {
	err error
}

// NewTxGenerationError creates a new tx generation error.
func NewTxGenerationError(err error) TxGenerationError {
	return TxGenerationError{
		err: err,
	}
}

// Error returns the error message.
func (t TxGenerationError) Error() string {
	return fmt.Errorf("error generating transaction: %w", t.err).Error()
}

// MsgGenerationError is an error that occurs when a transaction fails to be generated.
type MsgGenerationError struct {
	err error
}

// NewMsgGenerationError creates a new tx generation error.
func NewMsgGenerationError(err error) MsgGenerationError {
	return MsgGenerationError{
		err: err,
	}
}

// Error returns the error message.
func (t MsgGenerationError) Error() string {
	return fmt.Errorf("error generating sdk message: %w", t.err).Error()
}

// InvalidMarketError is an error that occurs when a market is invalid.
type InvalidMarketError struct {
	err error
}

// NewInvalidMarketError creates a new invalid market error.
func NewInvalidMarketError(err error) InvalidMarketError {
	return InvalidMarketError{
		err: err,
	}
}

// Error returns the error message.
func (i InvalidMarketError) Error() string {
	return fmt.Errorf("invalid market: %w", i.err).Error()
}
