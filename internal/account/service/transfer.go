package service

import (
	"errors"

	"github.com/matheuszzy/ledger/internal/account/domain"
)

// Transfer moves amount between accounts; nothing is saved unless both sides succeed.
func (s *Service) Transfer(origin, destination string, amount float64) (EventOutput, error) {
	if origin == "" || destination == "" {
		return EventOutput{}, domain.ErrInvalidID
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	from, err := s.repository.GetByID(origin)
	if err != nil {
		return EventOutput{}, err
	}
	if origin == destination {
		return EventOutput{}, domain.ErrSameAccount
	}
	to, err := s.repository.GetByID(destination)
	if errors.Is(err, domain.ErrNotFound) {
		to, err = domain.NewAccount(destination, 0)
	}
	if err != nil {
		return EventOutput{}, err
	}
	if err := from.GetBalance().Withdraw(amount); err != nil {
		return EventOutput{}, err
	}
	if err := to.GetBalance().Deposit(amount); err != nil {
		return EventOutput{}, err
	}
	s.repository.Save(from)
	s.repository.Save(to)
	return EventOutput{Origin: from, Destination: to}, nil
}
