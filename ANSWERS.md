# Answers

## Q1 — Validation Order

For a **FROZEN** wallet with `balance=100`, calling `Withdraw(-50)`:

| Version | First check | Returns |
|---------|-------------|---------|
| **A** | `status == StatusFrozen` | `ErrWalletFrozen` |
| **B** | `amount <= 0` | `ErrInvalidAmount` |

**Which is correct? Version B.**

Input validation should always precede state validation.  Here is why:

1. **Input errors are the caller's fault; state errors are the system's fault.**
   A negative amount is a programming error or a malformed request — the caller
   sent garbage data.  `ErrInvalidAmount` tells the caller to fix their call.
   `ErrWalletFrozen` tells the caller to change the wallet's state.  Returning
   `ErrWalletFrozen` for a negative amount misleads the caller: they might
   unfreeze the wallet and call again with `-50`, only to get the same invalid
   result.

2. **Deterministic validation regardless of runtime state.**  Input constraints
   (`amount > 0`) are invariants that do not change over time.  State checks
   depend on mutable data.  Validating inputs first makes the method's
   contract easier to reason about and test in isolation.

3. **Principle of least surprise.**  A caller who passes a nonsensical amount
   expects `ErrInvalidAmount`, full stop.  Any other error is surprising.

This is the order used in `Withdraw` and `Deposit` in `internal/domain/wallet.go`.

---

## Q2 — `InsufficientBalanceError` Fields

For `balance=100`, `Withdraw(150)`:

| Choice | `Required` | User-facing message example |
|--------|------------|-----------------------------|
| **A — requested amount** | `150` | *"Insufficient balance: you requested 150 cents but only 100 cents are available."* |
| **B — deficit** | `50` | *"Insufficient balance: you are short by 50 cents."* |

**Which is better UX? Choice A — the requested amount.**

Reasons:

* The user already knows what they tried to withdraw (they typed it).  Telling
  them the *deficit* forces mental arithmetic: "I need 50 more, so I have to
  add at least 50 to my wallet."

* The requested amount directly mirrors the original intent, making the message
  easy to generate at any layer without re-deriving context.

* The available balance (`100`) is already present in the struct, so the
  deficit (`50`) is trivially derivable by the display layer if needed:
  `deficit = insuf.Required - insuf.Available`.

* API consumers (mobile apps, frontend) can show a contextual UI: "You need
  50 more cents — top up your wallet?" using both fields without the server
  having to anticipate every display format.

`InsufficientBalanceError.Required` therefore holds the **requested withdrawal
amount**, as implemented in `internal/domain/errors.go`.

---

## Q3 — Why Domain Errors Must Not Be Wrapped in the Use-Case Layer

Wrapping a domain error with `fmt.Errorf("failed: %w", err)` in the use-case
layer is harmful for the following reasons:

1. **It breaks `errors.Is` / `errors.As` for callers of the use case.**
   `%w` does wrap the error (Go 1.13+), so `errors.Is` still works *today*,
   but the intent of the question is about *opaque* wrapping such as
   `errors.New("failed: " + err.Error())` — which throws away the chain.
   Even with `%w`, each wrapping layer adds noise and encourages callers to
   match on the *wrapped* string rather than the sentinel.

2. **Domain errors are part of the public contract.**  `ErrWalletFrozen`,
   `ErrInvalidAmount`, and `ErrInsufficientBalance` are the vocabulary the
   domain publishes.  The use-case layer is a thin orchestrator; it should
   propagate these errors as-is so that the transport layer (HTTP handler,
   gRPC interceptor) can translate them deterministically into status codes
   or response payloads.

3. **Wrapping hides structured information.**  `*InsufficientBalanceError`
   carries `Required` and `Available`.  Wrapping it in a generic string error
   destroys those fields, forcing the transport layer to parse error messages
   — fragile and brittle.

4. **Single responsibility.**  The use case should not decide how errors are
   presented; that is the transport layer's job.  Wrapping errors in the use
   case couples presentation concerns to business logic.

The correct pattern is to return domain errors directly and let the transport
layer map them:

```go
// use case
func (s *Service) Withdraw(ctx context.Context, id domain.WalletID, amount int64) error {
    w, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return err // infrastructure error, may be wrapped here if needed
    }
    return w.Withdraw(amount) // domain error — returned as-is
}

// HTTP handler
switch {
case errors.Is(err, domain.ErrWalletFrozen):
    http.Error(w, "wallet is frozen", http.StatusUnprocessableEntity)
case errors.Is(err, domain.ErrInvalidAmount):
    http.Error(w, "invalid amount", http.StatusBadRequest)
// ...
}
```

---

## Q4 — Differentiator: Why `Deposit` is Blocked on a FROZEN Wallet

**Decision:** `Deposit` returns `ErrWalletFrozen` when the wallet's status is
`StatusFrozen` (see `internal/domain/wallet.go`, line `if w.status == StatusFrozen`
inside `Deposit`).

**Why this approach:**
A frozen wallet represents an administrative hold — typically triggered by
fraud detection, compliance review, or a user-initiated lock.  The *purpose* of
freezing is to prevent the wallet from participating in financial flows until an
explicit review is completed.  Allowing deposits while blocking withdrawals
creates an **asymmetric state**:

- Funds can flow in but never out.
- An attacker or rogue process could inflate a frozen wallet's balance, making
  the post-review state harder to reconcile.
- Operations teams expect a frozen wallet to be *inert* — no surprise balance
  changes during review.

**The alternative:**
Allow deposits on a frozen wallet.  The argument is that accepting money from
a counterparty is "safe" because the funds cannot be moved out.  Some payment
systems use this model to avoid rejecting legitimate top-ups during a
temporary freeze.

**Trade-off:**

| | Block deposits on FROZEN | Allow deposits on FROZEN |
|---|---|---|
| **Consistency** | Wallet is fully inert | Asymmetric — in only |
| **Reconciliation** | Balance unchanged during review | Balance can change |
| **Counterparty UX** | Sender receives rejection and can retry | Sender succeeds silently |
| **Fraud surface** | Smaller | Larger (balance inflation) |
| **Operational clarity** | High | Medium |

For a generic payment domain entity without further context, the safer default
is to block all mutations on a frozen wallet and surface `ErrWalletFrozen`
consistently.  A future business requirement can always relax this constraint
with a targeted policy, whereas fixing silent balance inflation after the fact
is much harder.
