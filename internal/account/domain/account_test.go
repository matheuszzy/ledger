package domain

import (
	"errors"
	"testing"
)

// TestNewAccount covers valid creation and every validation failure.
func TestNewAccount(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		amount  float64
		wantErr error
	}{
		{"with initial balance", "100", 10, nil},
		{"with zero balance", "300", 0, nil},
		{"empty id", "", 10, ErrInvalidID},
		{"negative balance", "100", -1, ErrNegativeAmount},
		{"empty id wins over negative balance", "", -1, ErrInvalidID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := NewAccount(tt.id, tt.amount)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewAccount(%q, %v) error = %v, want %v", tt.id, tt.amount, err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if account != nil {
					t.Fatalf("NewAccount(%q, %v) = %+v, want nil on error", tt.id, tt.amount, account)
				}
				return
			}
			if account.GetID() != tt.id {
				t.Fatalf("GetID() = %q, want %q", account.GetID(), tt.id)
			}
			if account.GetBalance().GetAmount() != tt.amount {
				t.Fatalf("GetBalance().GetAmount() = %v, want %v", account.GetBalance().GetAmount(), tt.amount)
			}
		})
	}
}

// TestAccount_BalanceOperations checks that operations through GetBalance mutate the account itself.
func TestAccount_BalanceOperations(t *testing.T) {
	account, _ := NewAccount("100", 10)

	if err := account.GetBalance().Deposit(10); err != nil {
		t.Fatalf("Deposit error = %v, want nil", err)
	}
	if got := account.GetBalance().GetAmount(); got != 20 {
		t.Fatalf("balance after deposit = %v, want 20", got)
	}

	if err := account.GetBalance().Withdraw(5); err != nil {
		t.Fatalf("Withdraw error = %v, want nil", err)
	}
	if got := account.GetBalance().GetAmount(); got != 15 {
		t.Fatalf("balance after withdraw = %v, want 15", got)
	}
}

// TestAccount_Transfer replays the spec transfer: 15 from account 100 into a new account 300.
func TestAccount_Transfer(t *testing.T) {
	origin, _ := NewAccount("100", 15)
	destination, _ := NewAccount("300", 0)

	if err := origin.GetBalance().Withdraw(15); err != nil {
		t.Fatalf("origin Withdraw error = %v, want nil", err)
	}
	if err := destination.GetBalance().Deposit(15); err != nil {
		t.Fatalf("destination Deposit error = %v, want nil", err)
	}

	if got := origin.GetBalance().GetAmount(); got != 0 {
		t.Fatalf("origin balance = %v, want 0", got)
	}
	if got := destination.GetBalance().GetAmount(); got != 15 {
		t.Fatalf("destination balance = %v, want 15", got)
	}
}

// TestAccount_RejectedOperationsKeepBalance checks failed operations do not change the account.
func TestAccount_RejectedOperationsKeepBalance(t *testing.T) {
	account, _ := NewAccount("100", 10)

	if err := account.GetBalance().Deposit(-5); !errors.Is(err, ErrNegativeAmount) {
		t.Fatalf("Deposit(-5) error = %v, want ErrNegativeAmount", err)
	}
	if err := account.GetBalance().Withdraw(50); err == nil {
		t.Fatal("Withdraw(50) error = nil, want an error")
	}
	if got := account.GetBalance().GetAmount(); got != 10 {
		t.Fatalf("balance after rejected operations = %v, want 10", got)
	}
}
