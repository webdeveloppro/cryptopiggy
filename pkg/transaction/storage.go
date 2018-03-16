package transaction

import (
	"database/sql"
	"log"

	"github.com/pkg/errors"
)

// Storage general interface
type Storage interface {
	GetByWhere(string, ...interface{}) ([]Transaction, error)
	GetPricePerTransaction([]Transaction) error
}

// PGStorage for application working on postgresql database
type PGStorage struct {
	con *sql.DB
}

// NewStorage constructor
func NewStorage(con *sql.DB) *PGStorage {
	return &PGStorage{
		con: con,
	}
}

// GetByWhere execute sql query for transaction and find txin/txout data
func (pg *PGStorage) GetByWhere(sql string, val ...interface{}) ([]Transaction, error) {

	rows, err := pg.con.Query(sql, val)
	if err != nil {
		return make([]Transaction, 0), errors.Wrapf(err, "transaction: Cannot select transaction %s, %v", sql, val)
	}
	defer rows.Close()

	// ToDo
	// use len(rows) instead 0
	trans := make([]Transaction, 0)
	for rows.Next() {
		t := Transaction{}
		if err := rows.Scan(&t.ID, &t.BlockID, &t.Hash); err != nil {
			return trans, errors.Wrap(err, "transaction: Cannot scan for transaction")
		}

		txinRows, err := pg.con.Query(`SELECT
			ti.id, ti.amount, ti.prev_out, ti.size, ti.signature_script, ti.sequence, add.hash
			FROM txin as ti JOIN address as add on ti.address_id = add.id
			WHERE transaction_id = $1`, t.ID)
		if err != nil {
			return trans, errors.Wrap(err, "transaction: Cannot select for txint")
		}
		defer txinRows.Close()

		// ToDo
		// use len(rows) instead 0
		t.TxIns = make([]TxIn, 0)
		for txinRows.Next() {
			var txin TxIn
			if err := txinRows.Scan(
				&txin.ID,
				&txin.Amount,
				&txin.PrevOut,
				&txin.Size,
				&txin.SignatureScript,
				&txin.Sequence,
				&txin.Address); err != nil {
				return trans, errors.Wrap(err, "transaction: Cannot scan for txin")
			}
			t.TxIns = append(t.TxIns, txin)
		}

		txoutRows, err := pg.con.Query(`SELECT 
			id, val, pk_script
			FROM txout as tot
			WHERE transaction_id = $1`, t.ID)
		if err != nil {
			return trans, errors.Wrap(err, "transaction: Cannot select from txout")
		}
		defer txoutRows.Close()

		// ToDo
		// use len(rows) instead 0
		t.TxOuts = make([]TxOut, 0)
		for txoutRows.Next() {
			var txout TxOut
			if err := txoutRows.Scan(&txout.ID, &txout.Value, &txout.PkScript); err != nil {
				return trans, errors.Wrap(err, "transaction: Cannot scan for txout")
			}
			txout.getAddresses()
			t.TxOuts = append(t.TxOuts, txout)
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

		if err == sql.ErrNoRows {
			log.Printf("block: Cannot get bitcoin price, trans: %s, err: %v", t.Hash, err)
			t.Price = 0.00
		} else {
			return err
		}
	}
	return nil
}
