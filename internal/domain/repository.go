package domain

import "context"

// WalletMutation holds the fields that may be updated in a single persistence call.
// Using an explicit mutation type (rather than passing a full Wallet) prevents
// accidental full-object overwrites and makes the intent of each update explicit.
type WalletMutation struct {
	Balance *int64
	Status  *Status
}

// WalletRepository defines the persistence contract for Wallet aggregates.
// Implementations live in the infrastructure layer; the domain never imports them.
type WalletRepository interface {
	// FindByID returns the wallet for the given id.
	// Returns an error if the wallet does not exist or the query fails.
	FindByID(ctx context.Context, id WalletID) (*Wallet, error)

	// Save persists a new wallet.
	Save(ctx context.Context, wallet *Wallet) error

	// Update applies a partial mutation to an existing wallet.
	Update(ctx context.Context, id WalletID, mutation WalletMutation) error
}
