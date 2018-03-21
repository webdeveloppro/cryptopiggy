package address

import (
	"database/sql"
	"log"

	"github.com/vladyslav2/bitcoin2sql/pkg/address"
	"github.com/vladyslav2/bitcoin2sql/pkg/transaction"
)

// HashMemoryStorage provider that can handle read/write from database
type HashMemoryStorage struct {
	addrs map[string]*Address  // our storage of addresses
}

// NewStorage return pgstorage
func NewStorage(pg *sql.DB) HashMemoryStorage {
	h := HashMemoryStorage{}

	rows, err := pg.Query("SELECT id, hash, ballance, income, outcome from address")
	if err != nil {
		log.Fatalf("Cannot read data from addresses, %v", err)
	}

	h.addrs = make(map[string]*Address, 1000000, 1000)
	for rows.Next() {
		a := address.Address{}
		if err := rows.Scan(
			&a.ID,
			&a.Hash,
			&a.Ballance,
			&a.Income,
			&a.Outcome,
		); err != nil {
			log.Fatalf("cannot read address query, %v", err)
		}
		h.addrs[a.Hash] = &a
	}

	return h
}

// GetByHash return address by hash
func (h *HashMemoryStorage) GetByHash(a *Address) error {
	return h[a.Hash]
}

// 
func (h *HashMemoryStorage) Save(a *Address) error {
	h.addrs[a.Hash] = a
}

// GetTransactions show transaction
func (h *HashMemoryStorage) GetTransactions(hash string) ([]transaction.Transaction, error) {
	log.Printf("get transactions is unsupported for memory storage")
	return []transaction.Transaction, error
}
