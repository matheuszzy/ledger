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
		want    float64
		wantErr error
	}{
		{"with initial balance", "100", 10, 10, nil},
		{"with zero balance", "300", 0, 0, nil},
		{"fractional", "100", 20.5, 20.5, nil},
		{"rounds to cents", "100", 20.556, 20.56, nil},
		{"empty id", "", 10, 0, ErrInvalidID},
		{"negative balance", "100", -1, 0, ErrNegativeAmount},
		{"empty id wins over negative balance", "", -1, 0, ErrInvalidID},
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
			if account.ID() != tt.id {
				t.Fatalf("ID() = %q, want %q", account.ID(), tt.id)
			}
			if account.Balance() != tt.want {
				t.Fatalf("Balance() = %v, want %v", account.Balance(), tt.want)
			}
		})
	}
}

// TestAccount_BalanceOperations checks that deposits and withdrawals change the account itself.
func TestAccount_BalanceOperations(t *testing.T) {
	account, _ := NewAccount("100", 10)

	if err := account.Deposit(10); err != nil {
		t.Fatalf("Deposit error = %v, want nil", err)
	}
	if got := account.Balance(); got != 20 {
		t.Fatalf("balance after deposit = %v, want 20", got)
	}

	if err := account.Withdraw(5); err != nil {
		t.Fatalf("Withdraw error = %v, want nil", err)
	}
	if got := account.Balance(); got != 15 {
		t.Fatalf("balance after withdraw = %v, want 15", got)
	}
}

// TestAccount_Transfer replays the spec transfer: 15 from account 100 into a new account 300.
func TestAccount_Transfer(t *testing.T) {
	origin, _ := NewAccount("100", 15)
	destination, _ := NewAccount("300", 0)

	if err := origin.Withdraw(15); err != nil {
		t.Fatalf("origin Withdraw error = %v, want nil", err)
	}
	if err := destination.Deposit(15); err != nil {
		t.Fatalf("destination Deposit error = %v, want nil", err)
	}

	if got := origin.Balance(); got != 0 {
		t.Fatalf("origin balance = %v, want 0", got)
	}
	if got := destination.Balance(); got != 15 {
		t.Fatalf("destination balance = %v, want 15", got)
	}
}

// TestAccount_RejectedOperationsKeepBalance checks failed operations do not change the account.
func TestAccount_RejectedOperationsKeepBalance(t *testing.T) {
	account, _ := NewAccount("100", 10)

	if err := account.Deposit(-5); !errors.Is(err, ErrNegativeAmount) {
		t.Fatalf("Deposit(-5) error = %v, want ErrNegativeAmount", err)
	}
	if err := account.Withdraw(50); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("Withdraw(50) error = %v, want ErrInsufficientFunds", err)
	}
	if got := account.Balance(); got != 10 {
		t.Fatalf("balance after rejected operations = %v, want 10", got)
	}
}

// TestAccount_Deposit covers successful and rejected deposits.
func TestAccount_Deposit(t *testing.T) {
	tests := []struct {
		name    string
		start   float64
		amount  float64
		want    float64
		wantErr error
	}{
		{"into empty account", 0, 10, 10, nil},
		{"into existing account", 10, 10, 20, nil},
		{"fractional", 10, 0.5, 10.5, nil},
		{"sums without float error", 0.1, 0.2, 0.3, nil},
		{"rounds amount to cents", 10, 0.126, 10.13, nil},
		{"zero amount", 10, 0, 10, nil},
		{"negative amount", 10, -1, 10, ErrNegativeAmount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, _ := NewAccount("100", tt.start)

			err := account.Deposit(tt.amount)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Deposit(%v) error = %v, want %v", tt.amount, err, tt.wantErr)
			}
			if account.Balance() != tt.want {
				t.Fatalf("balance after Deposit(%v) = %v, want %v", tt.amount, account.Balance(), tt.want)
			}
		})
	}
}

// TestAccount_Withdraw covers successful withdrawals, including draining the balance to zero.
func TestAccount_Withdraw(t *testing.T) {
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
		{"subtracts without float error", 0.3, 0.1, 0.2},
		{"rounds amount to cents", 10, 0.126, 9.87},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, _ := NewAccount("100", tt.start)

			if err := account.Withdraw(tt.amount); err != nil {
				t.Fatalf("Withdraw(%v) error = %v, want nil", tt.amount, err)
			}
			if account.Balance() != tt.want {
				t.Fatalf("balance after Withdraw(%v) = %v, want %v", tt.amount, account.Balance(), tt.want)
			}
		})
	}
}

// TestAccount_Withdraw_NegativeAmount rejects negative withdrawals and leaves the account untouched.
func TestAccount_Withdraw_NegativeAmount(t *testing.T) {
	account, _ := NewAccount("100", 10)

	err := account.Withdraw(-1)

	if !errors.Is(err, ErrNegativeAmount) {
		t.Fatalf("Withdraw(-1) error = %v, want ErrNegativeAmount", err)
	}
	if account.Balance() != 10 {
		t.Fatalf("balance after rejected Withdraw = %v, want 10", account.Balance())
	}
}

// TestAccount_Withdraw_InsufficientFunds rejects overdrafts and leaves the account untouched.
func TestAccount_Withdraw_InsufficientFunds(t *testing.T) {
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
			account, _ := NewAccount("100", tt.start)

			if err := account.Withdraw(tt.amount); !errors.Is(err, ErrInsufficientFunds) {
				t.Fatalf("Withdraw(%v) from %v error = %v, want ErrInsufficientFunds", tt.amount, tt.start, err)
			}
			if account.Balance() != tt.start {
				t.Fatalf("balance after rejected Withdraw = %v, want %v", account.Balance(), tt.start)
			}
		})
	}
}

// TestAccount_APIScenario replays the deposit, withdraw and transfer amounts from the API spec.
func TestAccount_APIScenario(t *testing.T) {
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
