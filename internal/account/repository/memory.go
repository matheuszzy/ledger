package repository

import (
	"sync"

	"github.com/matheuszzy/ledger/internal/account/domain"
)

type Memory struct {
	mu       sync.RWMutex
	accounts map[string]*domain.Account
}

func NewMemory() *Memory {
	return &Memory{accounts: make(map[string]*domain.Account)}
}

func (m *Memory) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accounts = make(map[string]*domain.Account)
}

func (m *Memory) GetByID(id string) (*domain.Account, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	stored, existing := m.accounts[id]
	if !existing {
		return nil, domain.ErrNotFound
	}
	return clone(stored), nil
}

func (m *Memory) Save(account *domain.Account) {
	stored := clone(account)
	if stored == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accounts[stored.ID()] = stored
}

// clone returns a deep copy of the account, or nil when it is nil or not built by NewAccount.
func clone(account *domain.Account) *domain.Account {
	if account == nil {
		return nil
	}
	copied, err := domain.NewAccount(account.ID(), account.Balance())
	if err != nil {
		return nil
	}
	return copied
}
