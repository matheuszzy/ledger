package domain

import "math"

type Account struct {
	id      string
	balance float64
}

func NewAccount(id string, amount float64) (*Account, error) {
	if id == "" {
		return nil, ErrInvalidID
	}
	if amount < 0 {
		return nil, ErrNegativeAmount
	}
	return &Account{id: id, balance: roundCents(amount)}, nil
}

func (a *Account) ID() string {
	return a.id
}

func (a *Account) Balance() float64 {
	return a.balance
}

func (a *Account) Deposit(amount float64) error {
	if amount < 0 {
		return ErrNegativeAmount
	}
	a.balance = roundCents(a.balance + roundCents(amount))
	return nil
}

func (a *Account) Withdraw(amount float64) error {
	if amount < 0 {
		return ErrNegativeAmount
	}
	amount = roundCents(amount)
	if a.balance < amount {
		return ErrInsufficientFunds
	}
	a.balance = roundCents(a.balance - amount)
	return nil
}

// roundCents rounds an amount to two decimal places so float errors like 0.1+0.2 never reach the balance.
func roundCents(amount float64) float64 {
	cents := amount * 100
	if math.IsInf(cents, 0) {
		return amount
	}
	return math.Round(cents) / 100
}
