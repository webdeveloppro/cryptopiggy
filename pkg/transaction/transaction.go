package transaction

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// ErrNoTran Error message
var ErrNoTran = fmt.Errorf("No transaction found")

// Transaction holds transaction data and in/out array
type Transaction struct {
	ID         uint    `json:"id"`
	BlockID    uint    `json:"block_id"`
	Hash       string  `json:"hash"`
	HasWitness bool    `json:"has_witness"`
	Price      float32 `json:"price"`
	TxIns      []TxIn  `json:"txins"`
	TxOuts     []TxOut `json:"txouts"`
	Addresses  []uint
	storage    Storage
}

// New constructor
func New(storage Storage) *Transaction {
	return &Transaction{
		storage: storage,
	}
}

// Insert - create new record for current data, automarically fills ID value
func (t *Transaction) Insert() error {
	return t.storage.Insert(t)
}

// FindTransactions will search for transactions by giving key=val
func FindTransactions(reader Storage, key string, val interface{}) ([]Transaction, error) {
	var sql string

	if key == "address_hash" {
		sql = fmt.Sprintf(`SELECT
				id, block_id, hash, has_witness, txin, txout 
				FROM transaction where addresses@>'%d'
				ORDER BY id desc`, val)
		return reader.GetByWhere(sql)
	}

	sql = fmt.Sprintf(`SELECT
			id, block_id, hash, has_witness, txin, txout 
			FROM transaction as t
			WHERE %s = $1
			ORDER BY t.id desc`, key)

	return reader.GetByWhere(sql, val)
}

// FindTransaction will look for transaction where key=val
func FindTransaction(reader Storage, key string, val interface{}) (Transaction, error) {

	sql := fmt.Sprintf(`SELECT
			id, block_id, hash, has_witness, txin, txout 
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

// TxOutJSONB transform TxOut array for pg jsonb insert
func (t *Transaction) TxOutJSONB() string {
	if len(t.TxOuts) == 0 {
		return "[]"
	}

	pgQuery := ""
	for _, out := range t.TxOuts {
		pgQuery = fmt.Sprintf(`%s,{
  "val": %d,
  "pk_script": "%s"}`,
			pgQuery,
			out.Value,
			out.PkScript,
		)
	}
	pgQuery = strings.Replace(
		fmt.Sprintf("[%s]", pgQuery[1:len(pgQuery)]),
		"\n",
		"",
		-1,
	)
	return pgQuery
}

// TxInJSONB transform TxIn array for pg jsonb insert
func (t *Transaction) TxInJSONB() string {
	if len(t.TxIns) == 0 {
		return "[]"
	}

	pgQuery := ""
	for _, in := range t.TxIns {
		pgQuery = fmt.Sprintf(`%s,{
  "amount": %d,
  "address_id": %d,
  "address": "%s",
  "prev_out": "%s",
  "size": %d,
  "signature_script": "%s",
  "sequence": %d,
  "witness": "%s"}`,
			pgQuery,
			in.Amount,
			in.AddressID,
			in.Address,
			in.PrevOut,
			in.Size,
			in.SignatureScript,
			in.Sequence,
			in.Witness,
		)
	}
	pgQuery = strings.Replace(
		fmt.Sprintf("[%s]", pgQuery[1:len(pgQuery)]),
		"\n",
		"",
		-1,
	)
	return pgQuery
}

// AddressesJSONB transform Addresses array for pg jsonb insert
func (t *Transaction) AddressesJSONB() string {
	if len(t.Addresses) == 0 {
		return "[]"
	}

	pgQuery := ""
	for _, addr := range t.Addresses {
		pgQuery = fmt.Sprintf("%s,%d",
			pgQuery,
			addr,
		)
	}
	pgQuery = fmt.Sprintf("[%s]", pgQuery[1:len(pgQuery)])
	return pgQuery
}
