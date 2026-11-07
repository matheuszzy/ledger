package domain

type Balance struct {
	amount float64
}

func NewBalance(amount float64) (*Balance, error) {
	if amount < 0 {
		return nil, ErrNegativeAmount
	}
	return &Balance{amount: amount}, nil
}

func (b *Balance) GetAmount() float64 {
	return b.amount
}

func (b *Balance) Deposit(amount float64) error {
	if amount < 0 {
		return ErrNegativeAmount
	}
	b.amount += amount
	return nil
}

func (b *Balance) Withdraw(amount float64) error {
	if amount < 0 || b.amount < amount {
		return ErrNegativeAmount
	}
	b.amount -= amount
	return nil
}
