package transaction

import (
	"log"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
)

// Storage general interface
type Storage interface {
	Insert(*Transaction) error
	GetByWhere(string, ...interface{}) ([]Transaction, error)
	GetPricePerTransaction([]Transaction) error
}

// PGStorage for application working on postgresql database
type PGStorage struct {
	con *pgx.ConnPool
}

// NewStorage constructor
func NewStorage(con *pgx.ConnPool) *PGStorage {
	return &PGStorage{
		con: con,
	}
}

// Insert transaction to the database
func (pg *PGStorage) Insert(t *Transaction) error {
	sql := `
			INSERT INTO transaction
				(hash, block_id, has_witness, txin, txout, addresses)
			VALUES
				(
					$1,
					$2,
					$3,
					$4,
					$5,
					$6
				)
				RETURNING id`

	err := pg.con.QueryRow(sql,
		t.Hash,
		t.BlockID,
		t.HasWitness,
		t.TxInJSONB(),
		t.TxOutJSONB(),
		t.AddressesJSONB(),
	).Scan(&t.ID)

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"transaction_hash_key\"" &&
			t.Hash == "d5d27987d2a3dfc724e359870c6644b40e497bdc0589a033220fe15429d88599" {
			// https://bitcoin.stackexchange.com/questions/71918/why-does-transaction-d5d27987d2a3dfc724e359870c6644b40e497bdc0589a033220fe15429d
			return nil
		}

		return errors.Wrap(err, "insert transaction failed")
	}
	return nil
}

// GetByWhere execute sql query for transaction and find txin/txout data
func (pg *PGStorage) GetByWhere(sql string, val ...interface{}) ([]Transaction, error) {

	trans := make([]Transaction, 0)

	rows, err := pg.con.Query(sql, val...)
	if err != nil {
		return trans, err
	}

	for rows.Next() {
		t := Transaction{}
		if err := rows.Scan(
			// Transaction
			&t.ID, &t.BlockID, &t.Hash, &t.HasWitness,
			// txin
			&t.TxIns,
			// txout
			&t.TxOuts,
		); err != nil {
			if err == pgx.ErrNoRows {
				return trans, err
			}
			return trans, errors.Wrapf(err, "transaction: Cannot select transaction %s, %v", sql, val)
		}

		for i, in := range t.TxIns {
			if err := pg.con.QueryRow("SELECT hash FROM address WHERE id=$1",
				in.AddressID,
			).Scan(&t.TxIns[i].Address); err != nil {
				return trans, errors.Wrapf(err, "transaction: Cannot find address hash, transaction hash index:script - %s, %d:%s", t.Hash, i, in.SignatureScript)
			}
		}

		for i, out := range t.TxOuts {
			_, err := t.TxOuts[i].GetAddresses()
			if err != nil {
				return trans, errors.Wrapf(err, "transaction: Cannot decode pk_script, transaction hash index:script - %s, %d:%s", t.Hash, i, out.PkScript)
			}
		}
		trans = append(trans, t)
	}

	return trans, nil
}

// GetPricePerTransaction loop over all transactions and gets price per each transaction
func (pg *PGStorage) GetPricePerTransaction(trans []Transaction) error {
	for _, t := range trans {

		err := pg.con.QueryRow(`SELECT
			price
			FROM btc_price as bt JOIN block as bl on bt.created_at <= bl.created_at
			WHERE bl.id = $1 
			ORDER BY bt.created_at desc limit 1`,
			t.BlockID,
		).Scan(&t.Price)

		if err == pgx.ErrNoRows {
			log.Printf("block: Cannot get bitcoin price, trans: %s, err: %v", t.Hash, err)
			t.Price = 0.00
		} else {
			return err
		}
	}
	return nil
}
