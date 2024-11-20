package submitter

import (
	"fmt"
)

// CheckTxError is an error thrown from the transaction submitter when a check-tx fails.
type CheckTxError struct {
	log  string
	code uint32
}

func (e CheckTxError) Error() string {
	return fmt.Sprintf("check-tx error (code: %d): %s", e.code, e.log)
}

// NewCheckTxError creates a new CheckTxError.
func NewCheckTxError(log string, code uint32) CheckTxError {
	return CheckTxError{log: log, code: code}
}

// DeliverTxError is an error thrown from the transaction submitter when a deliver-tx fails.
type DeliverTxError struct {
	log  string
	code uint32
}

func (e DeliverTxError) Error() string {
	return fmt.Sprintf("deliver-tx error (code: %d): %s", e.code, e.log)
}

// NewDeliverTxError creates a new DeliverTxError.
func NewDeliverTxError(log string, code uint32) DeliverTxError {
	return DeliverTxError{log: log, code: code}
}

// TxTimeoutError is an error thrown from the transaction submitter when the
// submitter times out when polling for a transaction result.
type TxTimeoutError struct {
	err error
}

func (e TxTimeoutError) Error() string {
	return fmt.Sprintf("tx submitter timeout: %s", e.err)
}

// NewTxTimeoutError creates a new TxTimeoutError.
func NewTxTimeoutError(err error) TxTimeoutError {
	return TxTimeoutError{err: err}
}

// TxBroadcastError is an error thrown from the transaction submitter when the
// submitter fails to broadcast a transaction.
type TxBroadcastError struct {
	err error
}

func (e TxBroadcastError) Error() string {
	return fmt.Sprintf("tx broadcast error: %s", e.err)
}

// NewTxBroadcastError creates a new TxBroadcastError.
func NewTxBroadcastError(err error) TxBroadcastError {
	return TxBroadcastError{err: err}
}
