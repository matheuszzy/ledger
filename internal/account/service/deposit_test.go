package service

import (
	"errors"
	"testing"

	"github.com/matheuszzy/ledger/internal/account/domain"
)

func TestService_Deposit(t *testing.T) {
	svc, store := newService()

	result, err := svc.Deposit("100", 10)
	if err != nil {
		t.Fatalf("Deposit into new account error = %v, want nil", err)
	}
	if result.Destination.ID() != "100" || result.Destination.Balance() != 10 || result.Origin != nil {
		t.Fatalf("result = %+v, want destination {100 10} and no origin", result)
	}
	assertBalance(t, store, "100", 10)

	result, err = svc.Deposit("100", 10)
	if err != nil {
		t.Fatalf("Deposit into existing account error = %v, want nil", err)
	}
	if got := result.Destination.Balance(); got != 20 {
		t.Fatalf("result balance = %v, want 20", got)
	}
	assertBalance(t, store, "100", 20)
}

func TestService_Deposit_Errors(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		amount  float64
		wantErr error
	}{
		{"negative amount", "100", -1, domain.ErrNegativeAmount},
		{"empty id", "", 10, domain.ErrInvalidID},
	}
	for _, tt := range tests {
		t.Run(tt.name+" on new account", func(t *testing.T) {
			svc, store := newService()

			_, err := svc.Deposit(tt.id, tt.amount)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Deposit(%q, %v) error = %v, want %v", tt.id, tt.amount, err, tt.wantErr)
			}
			assertMissing(t, store, tt.id)
		})
	}
}

func TestService_Deposit_NegativeKeepsExistingBalance(t *testing.T) {
	svc, store := newService()
	seed(t, store, "100", 10)

	if _, err := svc.Deposit("100", -5); !errors.Is(err, domain.ErrNegativeAmount) {
		t.Fatalf("Deposit(-5) error = %v, want ErrNegativeAmount", err)
	}
	assertBalance(t, store, "100", 10)
}

func TestService_Concurrent_DepositsAreNotLost(t *testing.T) {
	const workers = 200
	svc, store := newService()

	runConcurrently(workers, func(int) {
		if _, err := svc.Deposit("100", 1); err != nil {
			t.Errorf("Deposit error = %v, want nil", err)
		}
	})

	assertBalance(t, store, "100", workers)
}
