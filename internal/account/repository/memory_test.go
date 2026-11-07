package repository

import (
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/matheuszzy/ledger/internal/account/domain"
)

// mustAccount builds an account from known-valid input; it panics instead of using t.Fatal so it is safe in goroutines.
func mustAccount(id string, amount float64) *domain.Account {
	account, err := domain.NewAccount(id, amount)
	if err != nil {
		panic(err)
	}
	return account
}

func TestMemory_GetByID_Missing(t *testing.T) {
	repo := NewMemory()

	account, err := repo.GetByID("1234")
	if !errors.Is(err, ErrNotFound) || account != nil {
		t.Fatalf("GetByID = (%v, %v), want (nil, ErrNotFound)", account, err)
	}
}

func TestMemory_SaveAndGet(t *testing.T) {
	repo := NewMemory()
	repo.Save(mustAccount("100", 10))
	repo.Save(mustAccount("100", 20.5))

	account, err := repo.GetByID("100")
	if err != nil {
		t.Fatalf("GetByID(100) error = %v, want nil", err)
	}
	if account.GetID() != "100" || account.GetBalance().GetAmount() != 20.5 {
		t.Fatalf("account = {%s %v}, want {100 20.5}", account.GetID(), account.GetBalance().GetAmount())
	}
}

func TestMemory_Save_Nil(t *testing.T) {
	repo := NewMemory()

	repo.Save(nil)

	if _, err := repo.GetByID(""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID after nil Save error = %v, want ErrNotFound", err)
	}
}

func TestMemory_Copies(t *testing.T) {
	repo := NewMemory()
	input := mustAccount("100", 10)
	repo.Save(input)

	if err := input.GetBalance().Deposit(989); err != nil {
		t.Fatalf("Deposit error = %v, want nil", err)
	}
	got, _ := repo.GetByID("100")
	if amount := got.GetBalance().GetAmount(); amount != 10 {
		t.Fatalf("mutating the Save argument changed stored balance to %v", amount)
	}

	if err := got.GetBalance().Deposit(545); err != nil {
		t.Fatalf("Deposit error = %v, want nil", err)
	}
	again, _ := repo.GetByID("100")
	if amount := again.GetBalance().GetAmount(); amount != 10 {
		t.Fatalf("mutating the GetByID result changed stored balance to %v", amount)
	}
}

func TestMemory_Reset(t *testing.T) {
	repo := NewMemory()
	repo.Save(mustAccount("100", 10))
	repo.Save(mustAccount("300", 15))

	repo.Reset()

	for _, id := range []string{"100", "300"} {
		if _, err := repo.GetByID(id); !errors.Is(err, ErrNotFound) {
			t.Fatalf("GetByID(%s) after Reset error = %v, want ErrNotFound", id, err)
		}
	}
	repo.Save(mustAccount("100", 1))
	if _, err := repo.GetByID("100"); err != nil {
		t.Fatalf("repository unusable after Reset: %v", err)
	}
}

func TestMemory_Concurrent(t *testing.T) {
	const workers = 50
	const iterations = 200
	repo := NewMemory()

	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			own := strconv.Itoa(worker)
			for i := 0; i < iterations; i++ {
				repo.Save(mustAccount(own, float64(i)))
				repo.Save(mustAccount("shared", float64(i)))
				if account, err := repo.GetByID("shared"); err == nil {
					_ = account.GetBalance().Deposit(1)
				}
				repo.GetByID(own)
				if worker == 0 && i%50 == 0 {
					repo.Reset()
				}
			}
		}(worker)
	}
	wg.Wait()
}

func TestMemory_Concurrent_DistinctIDs(t *testing.T) {
	const accounts = 500
	repo := NewMemory()

	var wg sync.WaitGroup
	for i := 0; i < accounts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			repo.Save(mustAccount(strconv.Itoa(i), float64(i)))
		}(i)
	}
	wg.Wait()

	for i := 0; i < accounts; i++ {
		account, err := repo.GetByID(strconv.Itoa(i))
		if err != nil || account.GetBalance().GetAmount() != float64(i) {
			t.Fatalf("account %d = (%v, %v), want balance %d", i, account, err, i)
		}
	}
}
