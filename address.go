package main

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

// Address Holds block data and table ID
type Address struct {
	ID           int           `json:"id" default:""`
	UpdatedAt    string        `json:"updated_at"`
	Hash         string        `json:"hash"`
	Transactions []Transaction `json:"transactions"`
	Income       int           `json:"income"`
	Outcome      int           `json:"outcome"`
	Ballance     int           `json:"ballance"`
	pg           *sql.DB
}

// NewAddress constructor for address structure
func NewAddress(pg *sql.DB) *Address {
	return &Address{
		pg: pg,
	}
}

// getByHash get object by hash
func (a *Address) getByHash(hash string) error {
	if err := a.pg.QueryRow(`
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
	); err != nil {
		if err.Error() == "sql: no rows in result set" {
			return fmt.Errorf("404")
		}
		return errors.Wrapf(err, "address:  get by hash failed %s", hash)
	}
	return nil
}

// getTransactions show transaction
func (a *Address) getTransactions() ([]Transaction, error) {

	var err error
	a.Transactions = make([]Transaction, 0, 0)

	if a.ID == 0 {
		return a.Transactions, nil
	}

	a.Transactions, err = FindTransactions(a.pg, "address_hash", a.Hash)
	if err != nil {
		return a.Transactions, errors.Wrap(err, "address: cannot get transactions")
	}

	return a.Transactions, nil
}

// getTransactions show transaction
func (a *Address) getPrices() error {

	if len(a.Transactions) == 0 {
		return nil
	}

	// FixMe
	// Create one sql query
	for i, tran := range a.Transactions {
		if err := a.pg.QueryRow(`SELECT
			price
			FROM btc_price as bt JOIN block as bl on bt.created_at <= bl.created_at
			WHERE bl.id = $1 
			ORDER BY bt.created_at desc limit 1`,
			tran.BlockID,
		).Scan(&a.Transactions[i].Price); err != nil {
			return errors.Wrapf(err, "address: Cannot find price for transaction %s", tran.Hash)
		}
	}
	return nil
}
