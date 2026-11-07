package domain

import (
	"errors"
	"testing"
)

// TestNewBalance covers valid and negative initial amounts.
func TestNewBalance(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		want    float64
		wantErr error
	}{
		{"positive", 10, 10, nil},
		{"zero", 0, 0, nil},
		{"fractional", 20.5, 20.5, nil},
		{"negative", -1, 0, ErrNegativeAmount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance, err := NewBalance(tt.amount)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewBalance(%v) error = %v, want %v", tt.amount, err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if balance != nil {
					t.Fatalf("NewBalance(%v) = %+v, want nil on error", tt.amount, balance)
				}
				return
			}
			if balance.GetAmount() != tt.want {
				t.Fatalf("GetAmount() = %v, want %v", balance.GetAmount(), tt.want)
			}
		})
	}
}

// TestBalance_Deposit covers successful and rejected deposits.
func TestBalance_Deposit(t *testing.T) {
	tests := []struct {
		name    string
		start   float64
		amount  float64
		want    float64
		wantErr error
	}{
		{"into empty balance", 0, 10, 10, nil},
		{"into existing balance", 10, 10, 20, nil},
		{"fractional", 10, 0.5, 10.5, nil},
		{"zero amount", 10, 0, 10, nil},
		{"negative amount", 10, -1, 10, ErrNegativeAmount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance, _ := NewBalance(tt.start)

			err := balance.Deposit(tt.amount)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Deposit(%v) error = %v, want %v", tt.amount, err, tt.wantErr)
			}
			if balance.GetAmount() != tt.want {
				t.Fatalf("balance after Deposit(%v) = %v, want %v", tt.amount, balance.GetAmount(), tt.want)
			}
		})
	}
}

// TestBalance_Withdraw covers successful withdrawals, including draining the balance to zero.
func TestBalance_Withdraw(t *testing.T) {
	tests := []struct {
		name   string
		start  float64
		amount float64
		want   float64
	}{
		{"partial", 20, 5, 15},
		{"exact balance", 15, 15, 0},
		{"zero amount", 10, 0, 10},
		{"fractional", 10, 0.5, 9.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance, _ := NewBalance(tt.start)

			if err := balance.Withdraw(tt.amount); err != nil {
				t.Fatalf("Withdraw(%v) error = %v, want nil", tt.amount, err)
			}
			if balance.GetAmount() != tt.want {
				t.Fatalf("balance after Withdraw(%v) = %v, want %v", tt.amount, balance.GetAmount(), tt.want)
			}
		})
	}
}

// TestBalance_Withdraw_NegativeAmount rejects negative withdrawals and leaves the balance untouched.
func TestBalance_Withdraw_NegativeAmount(t *testing.T) {
	balance, _ := NewBalance(10)

	err := balance.Withdraw(-1)

	if !errors.Is(err, ErrNegativeAmount) {
		t.Fatalf("Withdraw(-1) error = %v, want ErrNegativeAmount", err)
	}
	if balance.GetAmount() != 10 {
		t.Fatalf("balance after rejected Withdraw = %v, want 10", balance.GetAmount())
	}
}

// TestBalance_Withdraw_InsufficientFunds rejects overdrafts and leaves the balance untouched.
func TestBalance_Withdraw_InsufficientFunds(t *testing.T) {
	tests := []struct {
		name   string
		start  float64
		amount float64
	}{
		{"above balance", 10, 11},
		{"from empty balance", 0, 1},
		{"just above balance", 10, 10.01},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance, _ := NewBalance(tt.start)

			if err := balance.Withdraw(tt.amount); !errors.Is(err, ErrInsufficientFunds) {
				t.Fatalf("Withdraw(%v) from %v error = %v, want ErrInsufficientFunds", tt.amount, tt.start, err)
			}
			if balance.GetAmount() != tt.start {
				t.Fatalf("balance after rejected Withdraw = %v, want %v", balance.GetAmount(), tt.start)
			}
		})
	}
}

// TestBalance_APIScenario replays the deposit, withdraw and transfer amounts from the API spec.
func TestBalance_APIScenario(t *testing.T) {
	origin, _ := NewBalance(0)
	destination, _ := NewBalance(0)

	steps := []struct {
		name string
		do   func() error
		want float64
	}{
		{"create with initial deposit", func() error { return origin.Deposit(10) }, 10},
		{"deposit into existing", func() error { return origin.Deposit(10) }, 20},
		{"withdraw from existing", func() error { return origin.Withdraw(5) }, 15},
	}
	for _, step := range steps {
		if err := step.do(); err != nil {
			t.Fatalf("%s: error = %v, want nil", step.name, err)
		}
		if origin.GetAmount() != step.want {
			t.Fatalf("%s: balance = %v, want %v", step.name, origin.GetAmount(), step.want)
		}
	}

	if err := origin.Withdraw(15); err != nil {
		t.Fatalf("transfer withdraw: error = %v, want nil", err)
	}
	if err := destination.Deposit(15); err != nil {
		t.Fatalf("transfer deposit: error = %v, want nil", err)
	}
	if origin.GetAmount() != 0 || destination.GetAmount() != 15 {
		t.Fatalf("after transfer origin/destination = %v/%v, want 0/15", origin.GetAmount(), destination.GetAmount())
	}
}
