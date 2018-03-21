package address

import (
	"github.com/pkg/errors"
	"github.com/vladyslav2/bitcoin2sql/pkg/transaction"
)

// Address Holds block data and table ID
type Address struct {
	ID           int                       `json:"id" default:""`
	UpdatedAt    string                    `json:"updated_at"`
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
	return a.storage.Save(a)
}

// GetByHash Find Address by Hash
func (a *Address) GetByHash(hash string) error {
	a.Hash = hash
	return a.storage.GetByHash(a)
}

// GetTransactions show transaction
func (a *Address) GetTransactions() error {

	var err error
	a.Transactions = make([]transaction.Transaction, 0, 0)

	if a.ID == 0 {
		return nil
	}

	a.Transactions, err = a.storage.GetTransactions(a.Hash)
	if err != nil {
		return errors.Wrap(err, "address: cannot get transactions")
	}

	return nil
}
