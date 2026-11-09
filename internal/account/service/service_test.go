package service

import (
	"errors"
	"sync"
	"testing"

	"github.com/matheuszzy/ledger/internal/account/domain"
	"github.com/matheuszzy/ledger/internal/account/repository"
)

// newService builds a service over a real in-memory repository, returned so tests can inspect stored state.
func newService() (*Service, *repository.Memory) {
	store := repository.NewMemory()
	return New(store), store
}

// seed stores an account with the given balance.
func seed(t *testing.T, store *repository.Memory, id string, amount float64) {
	t.Helper()
	account, err := domain.NewAccount(id, amount)
	if err != nil {
		t.Fatalf("NewAccount(%q, %v) error = %v", id, amount, err)
	}
	store.Save(account)
}

// assertBalance checks the stored balance of an account.
func assertBalance(t *testing.T, store *repository.Memory, id string, want float64) {
	t.Helper()
	account, err := store.GetByID(id)
	if err != nil {
		t.Fatalf("account %q: error = %v, want it stored with balance %v", id, err, want)
	}
	if got := account.Balance(); got != want {
		t.Fatalf("account %q balance = %v, want %v", id, got, want)
	}
}

// assertMissing checks that no account is stored under id.
func assertMissing(t *testing.T, store *repository.Memory, id string) {
	t.Helper()
	if _, err := store.GetByID(id); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("account %q: error = %v, want ErrNotFound (account must not exist)", id, err)
	}
}

// runConcurrently runs fn from workers goroutines at once and waits for all of them.
func runConcurrently(workers int, fn func(worker int)) {
	var wg sync.WaitGroup
	for worker := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn(worker)
		}()
	}
	wg.Wait()
}
