package service

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/matheuszzy/ledger/internal/account/domain"
)

func TestService_Withdraw(t *testing.T) {
	svc, store := newService()
	seed(t, store, "100", 20)

	result, err := svc.Withdraw("100", 5)
	if err != nil {
		t.Fatalf("Withdraw error = %v, want nil", err)
	}
	if result.Origin.ID() != "100" || result.Origin.Balance() != 15 || result.Destination != nil {
		t.Fatalf("result = %+v, want origin {100 15} and no destination", result)
	}
	assertBalance(t, store, "100", 15)

	if _, err := svc.Withdraw("100", 15); err != nil {
		t.Fatalf("Withdraw of the full balance error = %v, want nil", err)
	}
	assertBalance(t, store, "100", 0)
}

func TestService_Withdraw_Errors(t *testing.T) {
	tests := []struct {
		name        string
		origin      string
		amount      float64
		wantErr     error
		wantBalance float64
	}{
		{"insufficient funds", "100", 11, domain.ErrInsufficientFunds, 10},
		{"negative amount", "100", -1, domain.ErrNegativeAmount, 10},
		{"empty id", "", 1, domain.ErrInvalidID, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, store := newService()
			seed(t, store, "100", 10)

			_, err := svc.Withdraw(tt.origin, tt.amount)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Withdraw(%q, %v) error = %v, want %v", tt.origin, tt.amount, err, tt.wantErr)
			}
			assertBalance(t, store, "100", tt.wantBalance)
		})
	}
}

func TestService_Withdraw_NonExistingAccount(t *testing.T) {
	svc, store := newService()

	_, err := svc.Withdraw("200", 10)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Withdraw error = %v, want ErrNotFound", err)
	}
	assertMissing(t, store, "200")
}

func TestService_Concurrent_WithdrawsNeverOverdraw(t *testing.T) {
	const balance, workers = 100, 200
	svc, store := newService()
	seed(t, store, "100", balance)

	var succeeded, rejected atomic.Int32
	runConcurrently(workers, func(int) {
		switch _, err := svc.Withdraw("100", 1); {
		case err == nil:
			succeeded.Add(1)
		case errors.Is(err, domain.ErrInsufficientFunds):
			rejected.Add(1)
		default:
			t.Errorf("Withdraw error = %v, want nil or ErrInsufficientFunds", err)
		}
	})

	if succeeded.Load() != balance || rejected.Load() != workers-balance {
		t.Fatalf("succeeded/rejected = %d/%d, want %d/%d", succeeded.Load(), rejected.Load(), balance, workers-balance)
	}
	assertBalance(t, store, "100", 0)
}
