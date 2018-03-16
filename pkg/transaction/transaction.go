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
				t.id, t.block_id, t.hash, 
				ti.id, ti.amount, ti.prev_out, ti.size, ti.signature_script, ti.sequence, add.hash,
				tot.id, tot.val, tot.pk_script
				FROM transaction as t JOIN txin as ti on t.id=ti.transaction_id 
				JOIN address as add on ti.address_id = add.id
				JOIN txout as tot on tot.transaction_id=t.id
				WHERE add.hash = $1
				ORDER BY t.id desc)
			UNION (SELECT
				t.id, t.block_id, t.hash, 
				ti.id, ti.amount, ti.prev_out, ti.size, ti.signature_script, ti.sequence, add.hash,
				tot.id, tot.val, tot.pk_script
				FROM transaction as t JOIN txin as ti on t.id=ti.transaction_id 
				JOIN address as add on ti.address_id = add.id
				JOIN txout as tot on tot.transaction_id=t.id
				JOIN address as add2 on tot.address_id = add2.id
				WHERE add2.hash = $1
				ORDER BY t.id desc)`)
	} else {
		sql = fmt.Sprintf(`SELECT
			t.id, t.block_id, t.hash, 
			ti.id, ti.amount, ti.prev_out, ti.size, ti.signature_script, ti.sequence, add.hash,
			tot.id, tot.val, tot.pk_script
			FROM transaction as t JOIN txin as ti on t.id=ti.transaction_id 
			JOIN address as add on ti.address_id = add.id
			JOIN txout as tot on tot.transaction_id=t.id
			WHERE %s = $1
			ORDER BY t.id desc`, key)
	}

	fmt.Println(sql, key, val)

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
