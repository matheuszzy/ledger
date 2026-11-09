package service

import "github.com/matheuszzy/ledger/internal/account/domain"

func (s *Service) Balance(id string) (float64, error) {
	if id == "" {
		return 0, domain.ErrInvalidID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	account, err := s.repository.GetByID(id)
	if err != nil {
		return 0, err
	}
	return account.Balance(), nil
}
