package domain_test

import (
	"errors"
	"testing"

	"github.com/enesconf/wallet-domain/internal/domain"
)

// helpers ─────────────────────────────────────────────────────────────────────

func activeWallet(balance int64) *domain.Wallet {
	return domain.NewWallet("w1", "owner1", balance)
}

func frozenWallet(balance int64) *domain.Wallet {
	w := activeWallet(balance)
	_ = w.Freeze()
	return w
}

// Deposit ─────────────────────────────────────────────────────────────────────

func TestDeposit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		wallet     *domain.Wallet
		amount     int64
		wantErr    error
		wantBal    int64
	}{
		{
			name:    "valid deposit increases balance",
			wallet:  activeWallet(100),
			amount:  50,
			wantBal: 150,
		},
		{
			name:    "zero amount returns ErrInvalidAmount",
			wallet:  activeWallet(100),
			amount:  0,
			wantErr: domain.ErrInvalidAmount,
			wantBal: 100,
		},
		{
			name:    "negative amount returns ErrInvalidAmount",
			wallet:  activeWallet(100),
			amount:  -1,
			wantErr: domain.ErrInvalidAmount,
			wantBal: 100,
		},
		{
			name:    "frozen wallet returns ErrWalletFrozen",
			wallet:  frozenWallet(100),
			amount:  50,
			wantErr: domain.ErrWalletFrozen,
			wantBal: 100,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.wallet.Deposit(tc.amount)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Deposit(%d) error = %v, want %v", tc.amount, err, tc.wantErr)
			}
			if tc.wallet.Balance() != tc.wantBal {
				t.Errorf("balance = %d, want %d", tc.wallet.Balance(), tc.wantBal)
			}
		})
	}
}

// Withdraw ────────────────────────────────────────────────────────────────────

func TestWithdraw(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wallet  *domain.Wallet
		amount  int64
		wantErr error
		wantBal int64
	}{
		{
			name:    "valid withdrawal decreases balance",
			wallet:  activeWallet(200),
			amount:  50,
			wantBal: 150,
		},
		{
			name:    "exact balance withdrawal succeeds",
			wallet:  activeWallet(100),
			amount:  100,
			wantBal: 0,
		},
		{
			name:    "zero amount returns ErrInvalidAmount",
			wallet:  activeWallet(100),
			amount:  0,
			wantErr: domain.ErrInvalidAmount,
			wantBal: 100,
		},
		{
			name:    "negative amount returns ErrInvalidAmount",
			wallet:  activeWallet(100),
			amount:  -10,
			wantErr: domain.ErrInvalidAmount,
			wantBal: 100,
		},
		{
			name:    "frozen wallet returns ErrWalletFrozen",
			wallet:  frozenWallet(200),
			amount:  50,
			wantErr: domain.ErrWalletFrozen,
			wantBal: 200,
		},
		{
			name:    "insufficient balance returns ErrInsufficientBalance via errors.Is",
			wallet:  activeWallet(100),
			amount:  500,
			wantErr: domain.ErrInsufficientBalance,
			wantBal: 100,
		},
		{
			// Input validation runs before state checks.
			// A negative amount on a frozen wallet returns ErrInvalidAmount, not ErrWalletFrozen.
			name:    "invalid amount on frozen wallet returns ErrInvalidAmount (input-first order)",
			wallet:  frozenWallet(100),
			amount:  -50,
			wantErr: domain.ErrInvalidAmount,
			wantBal: 100,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.wallet.Withdraw(tc.amount)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Withdraw(%d) error = %v, want %v", tc.amount, err, tc.wantErr)
			}
			if tc.wallet.Balance() != tc.wantBal {
				t.Errorf("balance = %d, want %d", tc.wallet.Balance(), tc.wantBal)
			}
		})
	}
}

// InsufficientBalanceError — errors.As ────────────────────────────────────────

func TestWithdraw_InsufficientBalance_ErrorsAs(t *testing.T) {
	t.Parallel()

	w := activeWallet(100)
	err := w.Withdraw(150)

	// errors.Is via Unwrap chain
	if !errors.Is(err, domain.ErrInsufficientBalance) {
		t.Fatal("errors.Is(err, ErrInsufficientBalance) must be true")
	}

	// errors.As for structured fields
	var insuf *domain.InsufficientBalanceError
	if !errors.As(err, &insuf) {
		t.Fatal("errors.As must unwrap to *InsufficientBalanceError")
	}
	if insuf.Required != 150 {
		t.Errorf("Required = %d, want 150", insuf.Required)
	}
	if insuf.Available != 100 {
		t.Errorf("Available = %d, want 100", insuf.Available)
	}
}

// Freeze ──────────────────────────────────────────────────────────────────────

func TestFreeze(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wallet  *domain.Wallet
		wantErr error
	}{
		{
			name:   "active wallet can be frozen",
			wallet: activeWallet(0),
		},
		{
			name:    "already-frozen wallet returns ErrWalletFrozen",
			wallet:  frozenWallet(0),
			wantErr: domain.ErrWalletFrozen,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.wallet.Freeze()
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Freeze() error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil && !tc.wallet.IsFrozen() {
				t.Error("wallet must be frozen after Freeze()")
			}
		})
	}
}

// InsufficientBalanceError.Error() ───────────────────────────────────────────

func TestInsufficientBalanceError_Error(t *testing.T) {
	t.Parallel()
	e := &domain.InsufficientBalanceError{Required: 150, Available: 100}
	msg := e.Error()
	if msg == "" {
		t.Error("Error() must return a non-empty string")
	}
}

// NewWallet ───────────────────────────────────────────────────────────────────

func TestNewWallet(t *testing.T) {
	t.Parallel()

	// Negative balance is clamped to zero.
	wNeg := domain.NewWallet("wneg", "owner0", -100)
	if wNeg.Balance() != 0 {
		t.Errorf("negative initial balance must be clamped to 0, got %d", wNeg.Balance())
	}

	w := domain.NewWallet("w99", "owner99", 500)
	if w.ID() != "w99" {
		t.Errorf("ID = %q, want w99", w.ID())
	}
	if w.OwnerID() != "owner99" {
		t.Errorf("OwnerID = %q, want owner99", w.OwnerID())
	}
	if w.Balance() != 500 {
		t.Errorf("Balance = %d, want 500", w.Balance())
	}
	if w.Status() != domain.StatusActive {
		t.Errorf("Status = %q, want ACTIVE", w.Status())
	}
}
