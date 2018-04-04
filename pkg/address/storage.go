package address

import (
	"github.com/jackc/pgx"
	"github.com/webdeveloppro/cryptopiggy/pkg/transaction"
)

// Storage is main interface for operations with Block
type Storage interface {
	GetByHash(*Address) error
	Insert(*Address) error
	Update(*Address) error
	GetTransactions(uint) ([]transaction.Transaction, error)
	GetAddresses(string, ...interface{}) ([]*Address, error)
}

// PGStorage provider that can handle read/write from database
type PGStorage struct {
	con *pgx.ConnPool
}

// NewStorage return pgstorage
func NewStorage(pg *pgx.ConnPool) PGStorage {
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

// Insert will create new address entity
func (pg *PGStorage) Insert(a *Address) error {
	var err error
	err = pg.con.QueryRow(`
		INSERT INTO address(hash, income, outcome, ballance)
		values($1, $2, $3, $4)
		RETURNING ID`,
		a.Hash,
		a.Income,
		a.Outcome,
		a.Ballance).Scan(
		&a.ID)
	return err
}

// Update will update address using id field
func (pg *PGStorage) Update(a *Address) error {
	var err error
	_, err = pg.con.Exec(`
			UPDATE address set ballance = $1, income = $2, outcome = $3
			WHERE id=$4`,
		a.Ballance,
		a.Income,
		a.Outcome,
		a.ID,
	)
	return err
}

// GetTransactions show transaction
func (pg *PGStorage) GetTransactions(id uint) ([]transaction.Transaction, error) {
	tranStorage := transaction.NewStorage(pg.con)
	return transaction.FindTransactions(tranStorage, "address_hash", id)
}

// GetAddresses return address according to sql query
func (pg *PGStorage) GetAddresses(sql string, args ...interface{}) ([]*Address, error) {
	sql = "SELECT id, updated_at, hash, income, outcome, ballance FROM address " + sql

	rows, err := pg.con.Query(sql, args)
	if err != nil {
		return nil, err
	}

	addresses := make([]*Address, 0)

	for rows.Next() {
		a := &Address{}
		if err := rows.Scan(
			&a.ID,
			&a.UpdatedAt,
			&a.Hash,
			&a.Income,
			&a.Outcome,
			&a.Ballance,
		); err != nil {
			return addresses, err
		}
		addresses = append(addresses, a)
	}
	return addresses, err
}
