package service

import (
	"errors"

	"github.com/matheuszzy/ledger/internal/account/domain"
)

// Deposit adds amount to the destination account, creating it when it does not exist.
func (s *Service) Deposit(destination string, amount float64) (EventOutput, error) {
	if destination == "" {
		return EventOutput{}, domain.ErrInvalidID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	account, err := s.repository.GetByID(destination)
	if errors.Is(err, domain.ErrNotFound) {
		account, err = domain.NewAccount(destination, 0)
	}
	if err != nil {
		return EventOutput{}, err
	}
	if err := account.Deposit(amount); err != nil {
		return EventOutput{}, err
	}
	s.repository.Save(account)
	return EventOutput{Destination: account}, nil
}
