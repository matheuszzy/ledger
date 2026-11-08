package service

import "github.com/matheuszzy/ledger/internal/account/domain"

func (s *Service) Balance(id string) (float64, error) {
	if id == "" {
		return 0, domain.ErrInvalidID
	}
	account, err := s.repository.GetByID(id)
	if err != nil {
		return 0, err
	}
	return account.GetBalance().GetAmount(), nil
}
