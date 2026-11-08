package service

import "github.com/matheuszzy/ledger/internal/account/domain"

// Withdraw removes amount from an existing origin account.
func (s *Service) Withdraw(origin string, amount float64) (EventOutput, error) {
	if origin == "" {
		return EventOutput{}, domain.ErrInvalidID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	account, err := s.repository.GetByID(origin)
	if err != nil {
		return EventOutput{}, err
	}
	if err := account.GetBalance().Withdraw(amount); err != nil {
		return EventOutput{}, err
	}
	s.repository.Save(account)
	return EventOutput{Origin: account}, nil
}
