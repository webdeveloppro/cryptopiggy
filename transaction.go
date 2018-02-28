package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/pkg/errors"
)

// Transaction holds transaction data and in/out array
type Transaction struct {
	ID         int     `json:"id"`
	BlockID    int     `json:"block_id"`
	Hash       string  `json:"hash"`
	HasWitness bool    `json:"has_witness"`
	Price      float32 `json:"price"`
	TxIns      []TxIn  `json:"txins"`
	TxOuts     []TxOut `json:"txouts"`
}

// TxIn transaction incoming data
type TxIn struct {
	ID              int    `json:"id"`
	Amount          int64  `json:"amount"`
	PrevOut         string `json:"prev_out"`
	Size            int    `json:"size"`
	SignatureScript string `json:"signature_script"`
	Sequence        int64  `json:"sequence"`
	Witness         string `json:"witness"`
	Address         string `json:"address"`
}

// TxOut transaction outcoming data
type TxOut struct {
	ID        int      `json:"id"`
	PkScript  string   `json:"pk_script"`
	Value     int64    `json:"value"`
	Addresses []string `json:"addresses"`
}

// getAddresses return addresses where money went
func (txOut *TxOut) getAddresses() ([]string, error) {
	if txOut.Addresses == nil {
		dst := make([]byte, hex.DecodedLen(len(txOut.PkScript)))
		_, err := hex.Decode(dst, []byte(txOut.PkScript))
		if err != nil {
			return []string{}, errors.Wrap(err, "block: Cannot convert hex string to bytes")
		}

		_, addresses, _, err := txscript.ExtractPkScriptAddrs(dst, &chaincfg.MainNetParams)
		if err != nil {
			return []string{}, errors.Wrap(err, fmt.Sprintf("Cannot extract pkScript %s", txOut.PkScript))
		}

		addrs := []string{}
		for _, a := range addresses {
			addrs = append(addrs, a.EncodeAddress())
		}
		txOut.Addresses = addrs
	}
	return txOut.Addresses, nil
}

// FindTransactions will search for transactions by giving key=val
func FindTransactions(pg *sql.DB, key string, val interface{}) ([]Transaction, error) {
	var sql string
	trans := make([]Transaction, 0)

	if key == "address_hash" {
		sql = fmt.Sprintf(`SELECT
			t.id, t.block_id, t.hash
			FROM transaction as t JOIN txout as txo ON t.id = txo.transaction_id 
			JOIN address as add ON txo.address_id=add.id WHERE add.hash=$1`)
	} else {
		sql = fmt.Sprintf(`SELECT
			t.id, t.block_id, t.hash
			FROM transaction as t
			WHERE %s = $1`, key)
	}

	rows, err := pg.Query(sql, val)
	if err != nil {
		return trans, errors.Wrap(err, "transaction: Cannot select for transaction")
	}

	for rows.Next() {
		t := Transaction{}
		if err := rows.Scan(&t.ID, &t.BlockID, &t.Hash); err != nil {
			return trans, errors.Wrap(err, "transaction: Cannot scan for transaction")
		}

		txinRows, err := pg.Query(`SELECT
			ti.id, ti.amount, ti.prev_out, ti.size, ti.signature_script, ti.sequence, add.hash
			FROM txin as ti JOIN address as add on ti.address_id = add.id
			WHERE transaction_id = $1`, t.ID)
		if err != nil {
			return trans, errors.Wrap(err, "transaction: Cannot select for txint")
		}

		t.TxIns = make([]TxIn, 0, 0)
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

		txoutRows, err := pg.Query(`SELECT 
			id, val, pk_script
			FROM txout as tot
			WHERE transaction_id = $1`, t.ID)
		if err != nil {
			return trans, errors.Wrap(err, "transaction: Cannot select from txout")
		}

		t.TxOuts = make([]TxOut, 0, 0)
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

// FindTransaction will look for transaction where key=val
func FindTransaction(pg *sql.DB, key string, val interface{}) (Transaction, error) {

	t := Transaction{}
	sql := fmt.Sprintf(`SELECT
			t.id, t.hash
			FROM transaction as t
			WHERE %s = $1`, key)

	if err := pg.QueryRow(sql, val).Scan(&t.ID, &t.Hash); err != nil {
		return t, errors.Wrap(err, "transaction: Cannot scan for transaction")
	}

	txinRows, err := pg.Query(`SELECT
			ti.id, ti.amount, ti.prev_out, ti.size, ti.signature_script, ti.sequence, add.hash
			FROM txin as ti JOIN address as add on ti.address_id = add.id
			WHERE transaction_id = $1`, t.ID)
	if err != nil {
		return t, errors.Wrap(err, "transaction: Cannot select for txint")
	}

	t.TxIns = make([]TxIn, 0, 0)
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
			return t, errors.Wrap(err, "transaction: Cannot scan for txin")
		}
		t.TxIns = append(t.TxIns, txin)
	}

	txoutRows, err := pg.Query(`SELECT 
			id, val, pk_script
			FROM txout as tot
			WHERE transaction_id = $1`, t.ID)
	if err != nil {
		return t, errors.Wrap(err, "transaction: Cannot select from txout")
	}

	t.TxOuts = make([]TxOut, 0, 0)
	for txoutRows.Next() {
		var txout TxOut
		if err := txoutRows.Scan(&txout.ID, &txout.Value, &txout.PkScript); err != nil {
			return t, errors.Wrap(err, "transaction: Cannot scan for txout")
		}
		txout.getAddresses()
		t.TxOuts = append(t.TxOuts, txout)
	}
	return t, nil
}
