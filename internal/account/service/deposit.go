package service

import (
	"errors"

	"github.com/matheuszzy/ledger/internal/account/domain"
	"github.com/matheuszzy/ledger/internal/account/repository"
)

func (s *Service) Deposit(destination string, amount float64) (EventOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	account, err := s.repository.GetByID(destination)
	if errors.Is(err, repository.ErrNotFound) {
		account, err = domain.NewAccount(destination, 0)
	}
	if err != nil {
		return EventOutput{}, err
	}
	if err := account.GetBalance().Deposit(amount); err != nil {
		return EventOutput{}, err
	}
	s.repository.Save(account)
	return EventOutput{Destination: account}, nil
}
