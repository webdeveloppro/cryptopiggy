package address

import (
	"database/sql"

	"github.com/vladyslav2/bitcoin2sql/pkg/transaction"
)

// Storage is main interface for operations with Block
type Storage interface {
	GetByHash(a *Address) error
	Insert(a *Address) error
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
func (pg *PGStorage) GetByHash(a *Address) error {
	err := pg.con.QueryRow(`
		SELECT id, updated_at, hash, income, outcome, ballance
		FROM address
		WHERE hash = $1
	`, a.Hash).Scan(
		&a.ID,
		&a.UpdatedAt,
		&a.Hash,
		&a.Income,
		&a.Outcome,
		&a.Ballance,
	)
	return err
}

// Save base on ID system will try to update object in database or create a new one
func (pg *PGStorage) Save(a *Address) error {
	var err error
	if a.ID == 0 {
		err = pg.con.QueryRow(`
			INSERT INTO address(hash, income, outcome, ballance)
			values($1, $2, $3, $4)
			RETURNING ID`,
			a.Hash,
			a.Income,
			a.Outcome,
			a.Ballance).Scan(
			&a.ID,
		)
	} else {
		_, err = pg.con.Exec(`
			UPDATE address set ballance = $1, income = $2, outcome = $3`,
			a.Ballance,
			a.Income,
			a.Outcome,
		)
	}
	return err
}

// GetTransactions show transaction
func (pg *PGStorage) GetTransactions(hash string) ([]transaction.Transaction, error) {
	tranStorage := transaction.NewStorage(pg.con)
	return transaction.FindTransactions(tranStorage, "address_hash", hash)
}
