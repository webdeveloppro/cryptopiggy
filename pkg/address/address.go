package address

import (
	"time"

	"github.com/pkg/errors"
	"github.com/webdeveloppro/cryptopiggy/pkg/transaction"
)

// Address Holds block data and table ID
type Address struct {
	ID           uint                      `json:"id" default:""`
	UpdatedAt    time.Time                 `json:"updated_at"`
	Hash         string                    `json:"hash"`
	Transactions []transaction.Transaction `json:"transactions"`
	Income       int64                     `json:"income"`
	Outcome      int64                     `json:"outcome"`
	Ballance     int64                     `json:"ballance"`
	storage      Storage
}

// New constructor for address structure
func New(storage Storage) *Address {
	a := Address{
		storage: storage,
	}
	return &a
}

// Save will create new record and ID
func (a *Address) Save() error {
	if a.ID == 0 {
		return a.storage.Insert(a)
	}
	return a.storage.Update(a)
}

// GetByHash Find Address by Hash
func (a *Address) GetByHash(hash string) error {
	a.Hash = hash
	return a.storage.GetByHash(a)
}

// GetTransactions show transaction
func (a *Address) GetTransactions() error {

	var err error

	if a.ID == 0 {
		return nil
	}

	a.Transactions, err = a.storage.GetTransactions(a.ID)
	if err != nil {
		return errors.Wrap(err, "address: cannot get transactions")
	}

	return nil
}

// Last10 show last 10 address
func Last10(storage Storage, order string) ([]*Address, error) {

	if order == "" {
		order = "ORDER BY id desc"
	}

	sql := "WHERE ballance != $1 ORDER BY " + order + " LIMIT 10"
	return storage.GetAddresses(sql, []int{10})
}

// MostRich10 show 10 richest addresses
func MostRich10(storage Storage) ([]*Address, error) {
	return Last10(storage, "ballance DESC")
}
