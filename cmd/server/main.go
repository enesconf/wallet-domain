package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/charlestest/wallet-domain/internal/domain"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// ── Domain demo ───────────────────────────────────────────────────────────
	// Demonstrates service-boundary logging as required by the spec.
	// In a real system this code lives in a use-case / service layer.

	w := domain.NewWallet("w1", "owner1", 10_000)

	// Successful deposit
	if err := w.Deposit(500); err != nil {
		logger.Warn("deposit rejected",
			"wallet_id", w.ID(),
			"amount_cents", 500,
			"reason", err.Error(),
			"error_type", fmt.Sprintf("%T", err),
		)
	} else {
		logger.Info("deposit completed",
			"wallet_id", w.ID(),
			"amount_cents", 500,
			"balance_cents", w.Balance(),
		)
	}

	// Successful withdrawal
	if err := w.Withdraw(200); err != nil {
		logger.Warn("withdraw rejected",
			"wallet_id", w.ID(),
			"amount_cents", 200,
			"reason", err.Error(),
			"error_type", fmt.Sprintf("%T", err),
		)
	} else {
		logger.Info("withdraw completed",
			"wallet_id", w.ID(),
			"amount_cents", 200,
			"balance_cents", w.Balance(),
		)
	}

	// Freeze the wallet then attempt a withdrawal
	if err := w.Freeze(); err != nil {
		logger.Warn("freeze rejected",
			"wallet_id", w.ID(),
			"reason", err.Error(),
			"error_type", fmt.Sprintf("%T", err),
		)
	} else {
		logger.Info("wallet frozen", "wallet_id", w.ID())
	}

	if err := w.Withdraw(100); err != nil {
		logger.Warn("withdraw rejected",
			"wallet_id", w.ID(),
			"amount_cents", 100,
			"reason", err.Error(),
			"error_type", fmt.Sprintf("%T", err),
		)
	}

	// Overdraft — demonstrates structured error logging
	w2 := domain.NewWallet("w2", "owner2", 50)
	if err := w2.Withdraw(300); err != nil {
		var insuf *domain.InsufficientBalanceError
		if errors.As(err, &insuf) {
			logger.Warn("withdraw rejected",
				"wallet_id", w2.ID(),
				"amount_cents", 300,
				"required_cents", insuf.Required,
				"available_cents", insuf.Available,
				"reason", err.Error(),
				"error_type", fmt.Sprintf("%T", err),
			)
		}
	}

	// ── HTTP server ───────────────────────────────────────────────────────────

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	addr := ":8080"
	logger.Info("server starting", "addr", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
