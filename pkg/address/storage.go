package address

import (
	"fmt"

	"github.com/jackc/pgx"
	"github.com/webdeveloppro/cryptopiggy/pkg/transaction"
)

// Storage is main interface for operations with Block
type Storage interface {
	GetByHash(*Address) error
	Save(*Address) error
	GetTransactions(uint) ([]transaction.Transaction, error)
	Last10(string) ([]*Address, error)
	MostRich() ([]*Address, error)
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
			UPDATE address set ballance = $1, income = $2, outcome = $3
			WHERE id=$4`,
			a.Ballance,
			a.Income,
			a.Outcome,
			a.ID,
		)
	}
	return err
}

// GetTransactions show transaction
func (pg *PGStorage) GetTransactions(id uint) ([]transaction.Transaction, error) {
	tranStorage := transaction.NewStorage(pg.con)
	return transaction.FindTransactions(tranStorage, "address_hash", id)
}

// Last10 show last 10 address
func (pg *PGStorage) Last10(order string) ([]*Address, error) {

	if order == "" {
		order = "id DESC"
	}

	addresses := make([]*Address, 0)
	rows, err := pg.con.Query(
		fmt.Sprintf(`
			SELECT id, updated_at, hash, income, outcome, ballance
			FROM address WHERE ballance != 0
			ORDER BY %s
			LIMIT 10`, order,
		),
	)

	if err != nil {
		return addresses, err
	}

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

// MostRich show 10 richest addresses
func (pg *PGStorage) MostRich() ([]*Address, error) {

	return pg.Last10("ballance DESC")
}
