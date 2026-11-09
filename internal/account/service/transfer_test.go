package service

import (
	"errors"
	"testing"

	"github.com/matheuszzy/ledger/internal/account/domain"
)

func TestService_Transfer(t *testing.T) {
	t.Run("to a new account", func(t *testing.T) {
		svc, store := newService()
		seed(t, store, "100", 15)

		result, err := svc.Transfer("100", "300", 15)
		if err != nil {
			t.Fatalf("Transfer error = %v, want nil", err)
		}
		if result.Origin.Balance() != 0 || result.Destination.Balance() != 15 {
			t.Fatalf("result balances = %v/%v, want 0/15", result.Origin.Balance(), result.Destination.Balance())
		}
		assertBalance(t, store, "100", 0)
		assertBalance(t, store, "300", 15)
	})

	t.Run("to an existing account", func(t *testing.T) {
		svc, store := newService()
		seed(t, store, "100", 20)
		seed(t, store, "300", 5)

		if _, err := svc.Transfer("100", "300", 8); err != nil {
			t.Fatalf("Transfer error = %v, want nil", err)
		}
		assertBalance(t, store, "100", 12)
		assertBalance(t, store, "300", 13)
	})
}

func TestService_Transfer_Errors(t *testing.T) {
	tests := []struct {
		name        string
		origin      string
		destination string
		amount      float64
		wantErr     error
	}{
		{"non-existing origin", "200", "300", 15, domain.ErrNotFound},
		{"insufficient funds", "100", "300", 11, domain.ErrInsufficientFunds},
		{"negative amount", "100", "300", -1, domain.ErrNegativeAmount},
		{"empty origin", "", "300", 1, domain.ErrInvalidID},
		{"empty destination", "100", "", 1, domain.ErrInvalidID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, store := newService()
			seed(t, store, "100", 10)

			_, err := svc.Transfer(tt.origin, tt.destination, tt.amount)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Transfer(%q, %q, %v) error = %v, want %v", tt.origin, tt.destination, tt.amount, err, tt.wantErr)
			}
			assertBalance(t, store, "100", 10)
			assertMissing(t, store, "200")
			assertMissing(t, store, "300")
		})
	}
}

func TestService_Transfer_FailureKeepsExistingDestination(t *testing.T) {
	svc, store := newService()
	seed(t, store, "100", 10)
	seed(t, store, "300", 5)

	if _, err := svc.Transfer("100", "300", 11); !errors.Is(err, domain.ErrInsufficientFunds) {
		t.Fatalf("Transfer error = %v, want ErrInsufficientFunds", err)
	}
	assertBalance(t, store, "100", 10)
	assertBalance(t, store, "300", 5)
}

func TestService_Transfer_SameAccount(t *testing.T) {
	svc, store := newService()
	seed(t, store, "100", 10)

	if _, err := svc.Transfer("100", "100", 5); !errors.Is(err, domain.ErrSameAccount) {
		t.Fatalf("Transfer to itself error = %v, want ErrSameAccount", err)
	}
	assertBalance(t, store, "100", 10)

	if _, err := svc.Transfer("200", "200", 5); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Transfer from non-existing account to itself error = %v, want ErrNotFound", err)
	}
	assertMissing(t, store, "200")
}

func TestService_Concurrent_TransfersConserveTotal(t *testing.T) {
	const start, workers = 1000, 200
	svc, store := newService()
	seed(t, store, "100", start)
	seed(t, store, "300", start)

	runConcurrently(workers, func(worker int) {
		origin, destination := "100", "300"
		if worker%2 == 1 {
			origin, destination = destination, origin
		}
		if _, err := svc.Transfer(origin, destination, 1); err != nil {
			t.Errorf("Transfer(%s -> %s) error = %v, want nil", origin, destination, err)
		}
	})

	first, _ := store.GetByID("100")
	second, _ := store.GetByID("300")
	if total := first.Balance() + second.Balance(); total != 2*start {
		t.Fatalf("total after concurrent transfers = %v, want %d", total, 2*start)
	}
	// Equal numbers of opposing transfers must cancel out exactly.
	assertBalance(t, store, "100", start)
	assertBalance(t, store, "300", start)
}
