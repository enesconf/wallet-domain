package domain

// WalletID is a typed string alias for wallet identifiers.
type WalletID string

// OwnerID is a typed string alias for owner identifiers.
type OwnerID string

// Status represents the lifecycle state of a Wallet.
type Status string

const (
	StatusActive Status = "ACTIVE"
	StatusFrozen Status = "FROZEN"
)

// Wallet is the domain aggregate for a payment wallet.
// All fields are unexported; mutation is only possible through methods.
type Wallet struct {
	id      WalletID
	ownerID OwnerID
	balance int64
	status  Status
}

// NewWallet creates a new active Wallet with the given identifiers and initial balance.
// A negative initial balance is clamped to zero to prevent inconsistent state.
func NewWallet(id WalletID, ownerID OwnerID, initialBalance int64) *Wallet {
	if initialBalance < 0 {
		initialBalance = 0
	}
	return &Wallet{
		id:      id,
		ownerID: ownerID,
		balance: initialBalance,
		status:  StatusActive,
	}
}

// ── Accessors ────────────────────────────────────────────────────────────────

func (w *Wallet) ID() WalletID      { return w.id }
func (w *Wallet) OwnerID() OwnerID  { return w.ownerID }
func (w *Wallet) Balance() int64    { return w.balance }
func (w *Wallet) Status() Status    { return w.status }
func (w *Wallet) IsFrozen() bool    { return w.status == StatusFrozen }

// ── Commands ─────────────────────────────────────────────────────────────────

// Deposit adds amount cents to the wallet balance.
//
// Validation order (inputs before state):
//  1. amount must be > 0  → ErrInvalidAmount
//  2. wallet must not be FROZEN → ErrWalletFrozen
//
// Design decision: Deposit is blocked on a FROZEN wallet.
// A frozen wallet implies an administrative hold; allowing deposits while
// withdrawals are blocked creates an asymmetric and potentially exploitable
// state. Full reasoning is in ANSWERS.md.
func (w *Wallet) Deposit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if w.status == StatusFrozen {
		return ErrWalletFrozen
	}
	w.balance += amount
	return nil
}

// Withdraw subtracts amount cents from the wallet balance.
//
// Validation order (inputs before state):
//  1. amount must be > 0  → ErrInvalidAmount
//  2. wallet must not be FROZEN → ErrWalletFrozen
//  3. balance must be sufficient → *InsufficientBalanceError (wraps ErrInsufficientBalance)
func (w *Wallet) Withdraw(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if w.status == StatusFrozen {
		return ErrWalletFrozen
	}
	if w.balance < amount {
		return &InsufficientBalanceError{
			Required:  amount,
			Available: w.balance,
		}
	}
	w.balance -= amount
	return nil
}

// Freeze transitions the wallet to the FROZEN state.
// Returns ErrWalletFrozen if the wallet is already frozen (idempotent guard).
func (w *Wallet) Freeze() error {
	if w.status == StatusFrozen {
		return ErrWalletFrozen
	}
	w.status = StatusFrozen
	return nil
}
