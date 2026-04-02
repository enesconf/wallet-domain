# wallet-domain

Backend Developer Assessment — Junior Level (Task 2)

A production-ready Go implementation of a `Wallet` domain entity for a payment system, demonstrating proper error handling, structured logging, and clean architecture.

---

## Project Structure

```
.
├── cmd/server/main.go              # Entry point: slog JSON logging + /healthz
├── internal/domain/
│   ├── errors.go                   # Sentinel + structured error types
│   ├── wallet.go                   # Wallet aggregate (private fields, typed IDs)
│   ├── repository.go               # WalletRepository interface + WalletMutation
│   └── wallet_test.go              # Table-driven unit tests (errors.Is + errors.As)
├── .github/workflows/ci.yml        # CI: test (race detector) + golangci-lint
├── .golangci.yml                   # Linter configuration
├── Makefile                        # build / test / lint / vet / tidy
├── REVIEW.md                       # Bug analysis + fix
└── ANSWERS.md                      # Q1–Q4 design decisions
```

---

## Getting Started

### Prerequisites

- Go 1.22+
- [golangci-lint](https://golangci-lint.run/usage/install/)

### Run

```bash
make run
```

Starts the HTTP server on `:8080`. Also prints structured JSON logs for demo wallet operations.

### Test

```bash
make test
```

Runs all tests with the race detector and prints a coverage report.

```
ok   github.com/charlestest/wallet-domain/internal/domain   100.0% coverage
```

### Lint

```bash
make lint
```

### All CI checks locally

```bash
make ci
```

---

## Health Check

```bash
curl http://localhost:8080/healthz
# {"status":"ok"}
```

---

## Domain Design

### Wallet Entity

| Field | Type | Description |
|-------|------|-------------|
| `id` | `WalletID` | Typed string identifier |
| `ownerID` | `OwnerID` | Typed string identifier |
| `balance` | `int64` | Balance in cents |
| `status` | `Status` | `ACTIVE` or `FROZEN` |

All fields are unexported. Mutation is only possible through domain methods.

### Methods

| Method | Description |
|--------|-------------|
| `Deposit(amount int64) error` | Adds cents to balance |
| `Withdraw(amount int64) error` | Subtracts cents from balance |
| `Freeze() error` | Transitions wallet to FROZEN state |

### Validation Order

Input validation always precedes state checks:

1. `amount > 0` → `ErrInvalidAmount`
2. `status != FROZEN` → `ErrWalletFrozen`
3. `balance >= amount` → `*InsufficientBalanceError`

This ensures callers receive the most actionable error regardless of wallet state.

### Error Types

```go
var (
    ErrInsufficientBalance = errors.New("insufficient balance")
    ErrInvalidAmount       = errors.New("invalid amount")
    ErrWalletFrozen        = errors.New("wallet is frozen")
)

// InsufficientBalanceError wraps ErrInsufficientBalance via Unwrap(),
// enabling both errors.Is and errors.As.
type InsufficientBalanceError struct {
    Required  int64 // requested withdrawal amount
    Available int64 // current balance
}
```

Usage:

```go
err := wallet.Withdraw(500)

if errors.Is(err, domain.ErrInsufficientBalance) {
    // handle insufficient balance
}

var insuf *domain.InsufficientBalanceError
if errors.As(err, &insuf) {
    fmt.Printf("need %d, have %d\n", insuf.Required, insuf.Available)
}
```

---

## Logging

Structured JSON logging via `log/slog` at the service boundary (never inside domain methods):

```json
{"time":"...","level":"INFO","msg":"deposit completed","wallet_id":"w1","amount_cents":500,"balance_cents":10500}
{"time":"...","level":"WARN","msg":"withdraw rejected","wallet_id":"w1","reason":"wallet is frozen","error_type":"*domain.WalletError"}
```

---

## CI

GitHub Actions runs two parallel jobs on every push and pull request:

- **Test** — `go vet` + `go test -race`
- **Lint** — `golangci-lint`

See [`.github/workflows/ci.yml`](.github/workflows/ci.yml).
