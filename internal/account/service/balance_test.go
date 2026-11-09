package service

import (
	"errors"
	"testing"

	"github.com/matheuszzy/ledger/internal/account/domain"
)

func TestService_Balance(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want float64
	}{
		{"positive balance", "100", 20},
		{"zero balance", "300", 0},
		{"fractional balance", "400", 10.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, store := newService()
			seed(t, store, tt.id, tt.want)

			got, err := svc.Balance(tt.id)

			if err != nil || got != tt.want {
				t.Fatalf("Balance(%q) = (%v, %v), want (%v, nil)", tt.id, got, err, tt.want)
			}
			assertBalance(t, store, tt.id, tt.want)
		})
	}
}

func TestService_Balance_Errors(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{"unknown account", "1234", domain.ErrNotFound},
		{"empty id", "", domain.ErrInvalidID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, store := newService()
			seed(t, store, "100", 20)

			got, err := svc.Balance(tt.id)

			if !errors.Is(err, tt.wantErr) || got != 0 {
				t.Fatalf("Balance(%q) = (%v, %v), want (0, %v)", tt.id, got, err, tt.wantErr)
			}
			assertBalance(t, store, "100", 20)
		})
	}
}

// TestService_Balance_HasNoSideEffects checks that reading never creates an account or changes a balance.
func TestService_Balance_HasNoSideEffects(t *testing.T) {
	svc, store := newService()
	seed(t, store, "100", 20)

	for range 3 {
		if _, err := svc.Balance("1234"); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("Balance(1234) error = %v, want ErrNotFound", err)
		}
		if got, err := svc.Balance("100"); err != nil || got != 20 {
			t.Fatalf("Balance(100) = (%v, %v), want (20, nil)", got, err)
		}
	}

	assertMissing(t, store, "1234")
	assertBalance(t, store, "100", 20)
}

// TestService_Concurrent_BalanceNeverSeesHalfTransfer reads both sides while transfers run and checks no money goes missing.
func TestService_Concurrent_BalanceNeverSeesHalfTransfer(t *testing.T) {
	const start = 2000
	svc, store := newService()
	seed(t, store, "100", start)
	seed(t, store, "300", 0)

	done := make(chan struct{})
	go func() {
		defer close(done)
		runConcurrently(4, func(int) {
			for range start / 4 {
				if _, err := svc.Transfer("100", "300", 1); err != nil {
					t.Errorf("Transfer(100 -> 300) error = %v, want nil", err)
					return
				}
			}
		})
	}()

	for {
		select {
		case <-done:
			assertBalance(t, store, "100", 0)
			assertBalance(t, store, "300", start)
			return
		default:
		}
		// Money only flows 100 -> 300, so reading 100 first can over-count but never under-count.
		origin, _ := svc.Balance("100")
		destination, _ := svc.Balance("300")
		if total := origin + destination; total < start {
			<-done
			t.Fatalf("read %v + %v = %v while transfers ran, want at least %d (saw a half-done transfer)", origin, destination, total, start)
		}
	}
}
