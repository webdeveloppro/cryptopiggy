package transaction

import (
	"fmt"

	"github.com/pkg/errors"
)

// ErrNoTran Error message
var ErrNoTran = fmt.Errorf("No transaction found")

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

// FindTransactions will search for transactions by giving key=val
func FindTransactions(reader Storage, key string, val interface{}) ([]Transaction, error) {
	var sql string

	if key == "address_hash" {
		sql = fmt.Sprintf(`(SELECT
			t.id, t.block_id, t.hash
			FROM transaction as t JOIN txout as txo ON t.id = txo.transaction_id 
			JOIN address as add ON txo.address_id=add.id WHERE add.hash=$1)
			UNION (SELECT 
				t.id, t.block_id, t.hash 
				FROM transaction as t JOIN txin as txo ON t.id=txo.transaction_id 
				JOIN address as add ON txo.address_id=add.id WHERE add.hash=$1)`)
	} else {
		sql = fmt.Sprintf(`SELECT
			t.id, t.block_id, t.hash
			FROM transaction as t
			WHERE %s = $1`, key)
	}

	return reader.GetByWhere(sql, val)
}

// FindTransaction will look for transaction where key=val
func FindTransaction(reader Storage, key string, val interface{}) (Transaction, error) {

	sql := fmt.Sprintf(`SELECT
			t.id, t.block_id, t.hash
			FROM transaction as t
			WHERE %s = $1`, key)

	trans, err := reader.GetByWhere(sql, val)
	if err != nil {
		return Transaction{}, errors.Wrapf(err, "Cannot find transaction: %s, %v", key, val)
	} else if len(trans) == 0 {
		return Transaction{}, ErrNoTran
	}

	return trans[0], nil
}

// GetPricePerTransaction loop over all transactions and gets price per each transaction
func GetPricePerTransaction(reader Storage, trans []Transaction) error {
	return reader.GetPricePerTransaction(trans)
}
