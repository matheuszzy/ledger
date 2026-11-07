package domain

type Account struct {
	id      string
	balance *Balance
}

func NewAccount(id string, amount float64) (*Account, error) {
	if id == "" {
		return nil, ErrInvalidID
	}
	balance, err := NewBalance(amount)
	if err != nil {
		return nil, err
	}
	return &Account{id: id, balance: balance}, nil
}

func (a *Account) GetID() string {
	return a.id
}

func (a *Account) GetBalance() *Balance {
	return a.balance
}
