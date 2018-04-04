package address_test

import (
	"testing"

	"github.com/jackc/pgx"

	"github.com/webdeveloppro/cryptopiggy/pkg/transaction"

	"github.com/webdeveloppro/cryptopiggy/pkg/address"
)

// FakeStorage layout to avoid database tests
type FakeStorage struct {
}

func (f *FakeStorage) GetByHash(a *address.Address) error {
	if a.Hash == "goodhash" {
		a.ID = 1
		a.Ballance = 10
	}
	return pgx.ErrNoRows
}

func (f *FakeStorage) Insert(a *address.Address) error {
	a.ID = 1
	return nil
}

func (f *FakeStorage) Update(a *address.Address) error {
	return nil
}

func (f *FakeStorage) GetTransactions(index uint) ([]transaction.Transaction, error) {
	return make([]transaction.Transaction, 0), nil
}

func (f *FakeStorage) GetAddresses(sql string, args ...interface{}) ([]*address.Address, error) {
	return make([]*address.Address, 0), nil
}

func (f *FakeStorage) MostRich() ([]*address.Address, error) {
	return make([]*address.Address, 0), nil
}

func TestSave(t *testing.T) {
	addr := address.New(&FakeStorage{})
	addr.Hash = "123"
	_ = addr.Save()

	if addr.ID != 1 {
		t.Errorf("Expected ID equal 1, got: %d", addr.ID)
	}
}

func TestGetByHash(t *testing.T) {
	addr := address.New(&FakeStorage{})
	err := addr.GetByHash("goodhash")
	if addr.ID != 1 {
		t.Errorf("Expected ID equal 1, got: %d", addr.ID)
	}

	err = addr.GetByHash("badhash")
	if err != pgx.ErrNoRows {
		t.Errorf("Error should be not nil for a bad hash")
	}
}

func TestGetTransactions(t *testing.T) {
	addr := address.New(&FakeStorage{})
	addr.GetTransactions()

	if addr.Transactions != nil {
		t.Errorf("Transactions should be nil")
	}

	addr.ID = 1
	if addr.Transactions != nil {
		t.Errorf("Transactions shouldn't be nil")
	}
}

func TestLast10(t *testing.T) {
	_, err := address.Last10(&FakeStorage{}, "")

	if err != nil {
		t.Errorf("Got error but should not")
	}
}

func TestMostRich10(t *testing.T) {
	_, err := address.MostRich10(&FakeStorage{})

	if err != nil {
		t.Errorf("Got error but should not")
	}
}
