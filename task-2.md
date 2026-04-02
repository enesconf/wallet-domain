# Backend Developer Assessment - Junior Level (Task 2)

## Instructions

1. Complete all tasks below
2. Push your solution to a **public GitHub repository**
3. Reply with your repository URL

---

## Task: Wallet Entity with Error Handling

Create a `Wallet` domain entity for a payment system with proper error handling.

### Requirements

**Fields:**
- `id` - WalletID
- `ownerID` - OwnerID
- `balance` - in cents (int64)
- `status` - ACTIVE, FROZEN

**Methods:**
- `Deposit(amount int64) error`
- `Withdraw(amount int64) error`
- `Freeze() error`

**Business Rules:**
- Amount must be positive (> 0) for both Deposit and Withdraw
- Cannot operate on FROZEN wallet
- Whether `Deposit` is allowed on a FROZEN wallet is intentionally left unspecified. Make your own decision and justify it in ANSWERS.md.
- Cannot withdraw more than available balance

**Error Types Required:**

```go
var (
    ErrInsufficientBalance = errors.New("insufficient balance")
    ErrInvalidAmount       = errors.New("invalid amount")
    ErrWalletFrozen        = errors.New("wallet is frozen")
)

// Structured error - must work with errors.Is()
type InsufficientBalanceError struct {
    Required  int64
    Available int64
}
```

---

## Buggy Code - Find the Problem

This test is failing. The code compiles but `errors.Is` never matches. Explain why in `REVIEW.md`:

```go
// errors.go
type WalletError struct {
    Code    string
    Message string
}

func (e WalletError) Error() string {
    return e.Message
}

var ErrInsufficientBalance = WalletError{Code: "E001", Message: "insufficient balance"}

// wallet.go
func (w *Wallet) Withdraw(amount int64) error {
    if w.balance < amount {
        return WalletError{
            Code:    "E001",
            Message: fmt.Sprintf("need %d, have %d", amount, w.balance),
        }
    }
    w.balance -= amount
    return nil
}

// wallet_test.go
func TestWithdraw_InsufficientBalance(t *testing.T) {
    w := NewWallet("w1", "owner1", 100)
    err := w.Withdraw(500)

    if !errors.Is(err, ErrInsufficientBalance) {
        t.Fatal("expected ErrInsufficientBalance")  // ALWAYS FAILS
    }
}
```

Your `REVIEW.md` must include:
1. Why `errors.Is` returns false
2. The exact fix (show corrected code)

---

## Questions - Answer in ANSWERS.md

**Q1:** Look at these two implementations. For a FROZEN wallet with balance=100, calling `Withdraw(-50)`:

```go
// Version A
func (w *Wallet) Withdraw(amount int64) error {
    if w.status == StatusFrozen { return ErrWalletFrozen }
    if amount <= 0 { return ErrInvalidAmount }
    // ...
}

// Version B
func (w *Wallet) Withdraw(amount int64) error {
    if amount <= 0 { return ErrInvalidAmount }
    if w.status == StatusFrozen { return ErrWalletFrozen }
    // ...
}
```

- What does Version A return?
- What does Version B return?
- Which is correct? Explain your reasoning.

**Q2:** Your `InsufficientBalanceError` has `Required` and `Available` fields. If balance=100 and someone tries to withdraw 150:

- Should `Required` be `150` (the requested amount) or `50` (the deficit)?
- Write the user-facing error message for each choice
- Which is better UX?

**Q3:** Why should domain errors NOT be wrapped with `fmt.Errorf("failed: %w", err)` in the usecase layer?

**Q4 (Differentiator):** Pick one specific decision you made in your own implementation and explain:
- Why did you choose this approach?
- What would the alternative have been?
- What is the trade-off between the two?

*Note: Your answer must reference a specific line or decision in your own code. Generic answers will not be accepted.*

---

## Monitoring & Observability

Add basic observability to your submission:

**1. Structured Logging (`main.go`)**

Use the standard `log/slog` package:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("deposit completed", "wallet_id", walletID, "amount_cents", amount)
logger.Warn("withdraw rejected", "wallet_id", walletID, "reason", err.Error())
```

**Rules:**
- Use `slog.NewJSONHandler` — never `fmt.Println` or `log.Printf`
- Log domain operation outcomes at the service/handler boundary, not inside domain methods
- Include `wallet_id` and relevant fields in every log entry
- Log errors with their type: `"error_type", fmt.Sprintf("%T", err)`

**2. Health Check Endpoint**

```go
http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
})
```

---

## Repository Structure

```
your-repo/
├── cmd/
│   └── server/
│       └── main.go              # Entry point: slog + /healthz
├── internal/
│   └── domain/
│       ├── errors.go            # Sentinel + structured error types
│       ├── wallet.go            # Wallet aggregate: private fields, accessor methods
│       ├── repository.go        # WalletRepository interface + WalletMutation type
│       └── wallet_test.go       # Table-driven unit tests (errors.Is + errors.As)
├── .github/
│   └── workflows/
│       └── ci.yml
├── .golangci.yml
├── Makefile
├── go.mod
├── REVIEW.md
└── ANSWERS.md
```

---

## Evaluation

Your submission will be evaluated against our engineering standards document. Key areas:
- Proper error types with Is() method implementation
- Validation order (inputs before state checks)
- Domain errors returned as-is, not wrapped
- Test coverage including errors.Is and errors.As usage
- Structured errors with context where needed
- Structured JSON logging with `log/slog` at service boundary (not in domain)
- `/healthz` endpoint returning JSON status
- Standard Go project layout (cmd/, internal/)
- Table-driven tests with t.Parallel()
- Makefile with build/test/lint targets
- CI workflow
- golangci-lint configuration
