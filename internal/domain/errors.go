package domain

import (
	"errors"
	"fmt"
)

// Sentinel errors — use errors.Is() to match these.
var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrWalletFrozen        = errors.New("wallet is frozen")
)

// InsufficientBalanceError is a structured error that wraps ErrInsufficientBalance
// so it satisfies both errors.Is(err, ErrInsufficientBalance) and errors.As().
//
// Required holds the requested withdrawal amount; Available holds the current balance.
// This lets callers build precise user-facing messages without string parsing.
type InsufficientBalanceError struct {
	Required  int64
	Available int64
}

// Error implements the error interface.
func (e *InsufficientBalanceError) Error() string {
	return fmt.Sprintf(
		"insufficient balance: required %d cents, available %d cents",
		e.Required, e.Available,
	)
}

// Unwrap chains to the sentinel so errors.Is(err, ErrInsufficientBalance) == true.
func (e *InsufficientBalanceError) Unwrap() error {
	return ErrInsufficientBalance
}
