package block

import (
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/vladyslav2/bitcoin2sql/pkg/transaction"
)

// Block Holds block data and table ID
type Block struct {
	ID             uint                      `json:"id" default:""`
	Bits           uint32                    `json:"bits"`
	Height         int32                     `json:"height"`
	Nonce          uint32                    `json:"nonce"`
	Version        int32                     `json:"version"`
	HashPrevBlock  string                    `json:"hash_prev_block"`
	HashMerkleRoot string                    `json:"hash_merkle_root"`
	CreatedAt      time.Time                 `json:"created_at"`
	Hash           string                    `json:"hash"`
	Transactions   []transaction.Transaction `json:"transactions"`
	Price          float32                   `json:"price"`
	storage        Storage
}

// New constructor for block structure
func New(storage Storage) *Block {
	b := Block{
		storage: storage,
	}
	b.Transactions = make([]transaction.Transaction, 0, 0)
	return &b
}

// GetTransactions show transaction
func (b *Block) GetTransactions() ([]transaction.Transaction, error) {
	var err error

	if b.ID == 0 {
		return b.Transactions, nil
	}

	b.Transactions, err = b.storage.getTransactions(b.ID)
	if err != nil {
		return b.Transactions, errors.Wrap(err, "block: cannot get transactions")
	}

	return b.Transactions, nil
}

// GetPrice return decimal price of bitcoin on the moment when block was created
func (b *Block) GetPrice() (err error) {
	b.Price, err = b.storage.getPrice(b.CreatedAt)
	if err != nil {
		log.Printf("block: Cannot get bitcoin price, block: %d, timestamp: %s, err: %v", b.ID, b.CreatedAt, err)
		return err
	}
	return err
}

// Insert will create new record for current block
func (b *Block) Insert() error {
	return b.storage.Insert(b)
}

// Last10 return last 10 blocks
func Last10(storage Storage) ([]Block, error) {
	return storage.Last10()
}
