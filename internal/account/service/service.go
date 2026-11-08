package service

import (
	"sync"

	"github.com/matheuszzy/ledger/internal/account/domain"
)

type Repository interface {
	GetByID(id string) (*domain.Account, error)
	Save(account *domain.Account)
	Reset()
}

// EventOutput holds the accounts affected by an event.
type EventOutput struct {
	Origin      *domain.Account
	Destination *domain.Account
}

type Service struct {
	mu         sync.Mutex
	repository Repository
}

func New(store Repository) *Service {
	return &Service{repository: store}
}
