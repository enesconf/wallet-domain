# Code Review — Bug Analysis

## The Failing Test

```go
func TestWithdraw_InsufficientBalance(t *testing.T) {
    w := NewWallet("w1", "owner1", 100)
    err := w.Withdraw(500)

    if !errors.Is(err, ErrInsufficientBalance) {
        t.Fatal("expected ErrInsufficientBalance") // ALWAYS FAILS
    }
}
```

## Why `errors.Is` Returns False

`errors.Is` traverses the **Unwrap chain** of an error, comparing each node to
the target with `==`.  Two values are equal under `==` only when they are the
same value in memory (for non-comparable types) or have identical fields (for
comparable struct types).

The sentinel is declared as:

```go
var ErrInsufficientBalance = WalletError{Code: "E001", Message: "insufficient balance"}
```

`Withdraw` returns a **new** `WalletError` value:

```go
return WalletError{
    Code:    "E001",
    Message: fmt.Sprintf("need %d, have %d", amount, w.balance),
}
```

There are two independent reasons this comparison fails:

1. **Different `Message` values.**  The sentinel has `"insufficient balance"`;
   the returned error has `"need 500, have 100"`.  Even though both have the
   same `Code`, struct equality requires every field to match.

2. **No `Unwrap` method.**  `WalletError` does not implement `Unwrap() error`,
   so `errors.Is` cannot walk any chain — it can only compare the top-level
   value, which already differs.

Because both conditions fail simultaneously, `errors.Is` returns `false` and
the test always reaches `t.Fatal`.

## The Exact Fix

There are two equivalent correct approaches.

### Option A — pointer sentinel with `Is` method (recommended)

Define the sentinel as a pointer and give the struct an `Is` method that
matches only on `Code`:

```go
// errors.go

type WalletError struct {
    Code    string
    Message string
}

func (e *WalletError) Error() string { return e.Message }

// Is lets errors.Is match any *WalletError with the same Code,
// regardless of the dynamic Message content.
func (e *WalletError) Is(target error) bool {
    t, ok := target.(*WalletError)
    if !ok {
        return false
    }
    return e.Code == t.Code
}

var ErrInsufficientBalance = &WalletError{Code: "E001", Message: "insufficient balance"}

// wallet.go
func (w *Wallet) Withdraw(amount int64) error {
    if w.balance < amount {
        return &WalletError{
            Code:    "E001",
            Message: fmt.Sprintf("need %d, have %d", amount, w.balance),
        }
    }
    w.balance -= amount
    return nil
}
```

`errors.Is(err, ErrInsufficientBalance)` now calls `err.Is(ErrInsufficientBalance)`,
which compares `Code` fields → both are `"E001"` → returns `true`.

### Option B — structured error with `Unwrap` (used in this submission)

Define a rich error type that wraps the sentinel via `Unwrap`:

```go
type InsufficientBalanceError struct {
    Required  int64
    Available int64
}

func (e *InsufficientBalanceError) Error() string {
    return fmt.Sprintf("insufficient balance: required %d cents, available %d cents",
        e.Required, e.Available)
}

// Unwrap chains to the sentinel so errors.Is traversal succeeds.
func (e *InsufficientBalanceError) Unwrap() error {
    return ErrInsufficientBalance
}

var ErrInsufficientBalance = errors.New("insufficient balance")
```

`Withdraw` returns `&InsufficientBalanceError{...}`.
`errors.Is` unwraps it, finds `ErrInsufficientBalance`, compares with `==` →
the same pointer → `true`.

This is the approach used in `internal/domain/errors.go` and
`internal/domain/wallet.go` in this submission, because it also enables
`errors.As` for structured field access.
