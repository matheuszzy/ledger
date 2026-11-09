package domain

import "math"

type Balance struct {
	amount float64
}

func NewBalance(amount float64) (*Balance, error) {
	if amount < 0 {
		return nil, ErrNegativeAmount
	}
	return &Balance{amount: roundCents(amount)}, nil
}

func (b *Balance) GetAmount() float64 {
	return b.amount
}

func (b *Balance) Deposit(amount float64) error {
	if amount < 0 {
		return ErrNegativeAmount
	}
	b.amount = roundCents(b.amount + roundCents(amount))
	return nil
}

func (b *Balance) Withdraw(amount float64) error {
	if amount < 0 {
		return ErrNegativeAmount
	}
	amount = roundCents(amount)
	if b.amount < amount {
		return ErrInsufficientFunds
	}
	b.amount = roundCents(b.amount - amount)
	return nil
}

func roundCents(amount float64) float64 {
	cents := amount * 100
	if math.IsInf(cents, 0) {
		return amount
	}
	return math.Round(cents) / 100
}
