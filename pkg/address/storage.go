package address

import (
	"database/sql"

	"github.com/vladyslav2/bitcoin2sql/pkg/transaction"
)

// Storage is main interface for operations with Block
type Storage interface {
	GetByHash(string) (*Address, error)
	GetTransactions(string) ([]transaction.Transaction, error)
}

// PGStorage provider that can handle read/write from database
type PGStorage struct {
	con *sql.DB
}

// NewStorage return pgstorage
func NewStorage(pg *sql.DB) PGStorage {
	return PGStorage{
		con: pg,
	}
}

// GetByHash return address by hash
func (pg *PGStorage) GetByHash(hash string) (*Address, error) {
	a := Address{storage: pg}
	err := pg.con.QueryRow(`
		SELECT id, updated_at, hash, income, outcome, ballance
		FROM address
		WHERE hash = $1
	`, hash).Scan(
		&a.ID,
		&a.UpdatedAt,
		&a.Hash,
		&a.Income,
		&a.Outcome,
		&a.Ballance,
	)
	return &a, err
}

// GetTransactions show transaction
func (pg *PGStorage) GetTransactions(hash string) ([]transaction.Transaction, error) {
	tranStorage := transaction.NewStorage(pg.con)
	return transaction.FindTransactions(tranStorage, "address_hash", hash)
}
