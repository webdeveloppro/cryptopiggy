package transaction

import (
	"database/sql"
	"log"

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
	con *sql.DB
}

// NewStorage constructor
func NewStorage(con *sql.DB) *PGStorage {
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

	rows, err := pg.con.Query(sql, val...)
	if err != nil {
		return make([]Transaction, 0), errors.Wrapf(err, "transaction: Cannot select transaction %s, %v", sql, val)
	}
	defer rows.Close()

	// 6 Steps to get data from sql to golang structures
	// First we need to understand where if end of one transaction and begin of the new one
	// Lets keep transaction/txin and txout indexes
	tIdx := -1
	tinIdx := -1
	toutIdx := -1

	// ToDo
	// use len(rows) instead 0
	trans := make([]Transaction, 0)

	for rows.Next() {
		// Get data to the temporary location
		t := Transaction{}
		txin := TxIn{}
		txout := TxOut{}

		if err := rows.Scan(
			// Transaction
			&t.ID, &t.BlockID, &t.Hash,
			// txin
			&txin.ID, &txin.Amount, &txin.PrevOut,
			&txin.Size, &txin.SignatureScript, &txin.Sequence, &txin.Address,
			// txout
			&txout.ID, &txout.Value, &txout.PkScript,
		); err != nil {
			return trans, errors.Wrap(err, "Transaction: Cannot scan for transaction")
		}

		//txout.getAddresses()

		// if t have different ID - that means we got new transaction
		if tIdx == -1 || trans[tIdx].ID != t.ID {
			trans = append(trans, t)
			tIdx++

			trans[tIdx].TxIns = make([]TxIn, 0)
			trans[tIdx].TxOuts = make([]TxOut, 0)
			tinIdx = -1
			toutIdx = -1
		}

		// if txin have different ID - we got a new txin
		if tinIdx == -1 || trans[tIdx].TxIns[tinIdx].ID != txin.ID {
			trans[tIdx].TxIns = append(trans[tIdx].TxIns, txin)
			tinIdx++
		}

		// if txin have different ID - we got a new txin
		if toutIdx == -1 || trans[tIdx].TxOuts[toutIdx].ID != txout.ID {
			trans[tIdx].TxOuts = append(trans[tIdx].TxOuts, txout)
			toutIdx++
		}

		/*

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
		*/
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
