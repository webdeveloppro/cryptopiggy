package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pkg/errors"
)

// Block Holds block data and table ID
type Block struct {
	ID             int           `json:"id" default:""`
	Bits           int           `json:"bits"`
	Height         int           `json:"height"`
	Nonce          int64         `json:"nonce"`
	Version        int           `json:"version"`
	HashPrevBlock  string        `json:"hash_prev_block"`
	HashMerkleRoot string        `json:"hash_merkle_root"`
	CreatedAt      string        `json:"created_at"`
	Hash           string        `json:"hash"`
	Transactions   []Transaction `json:"transactions"`
	Price          float32       `json:"price"`
	pg             *sql.DB
}

// getByHash get object by hash
func (b *Block) getByHash(hash string) error {
	if err := b.pg.QueryRow(`
		SELECT id, bits, height, nonce, version, hash_prev_block, hash_merkle_root, created_at, hash
		FROM block
		WHERE hash = $1
	`, hash).Scan(
		&b.ID,
		&b.Bits,
		&b.Height,
		&b.Nonce,
		&b.Version,
		&b.HashPrevBlock,
		&b.HashMerkleRoot,
		&b.CreatedAt,
		&b.Hash,
	); err != nil {
		if err.Error() == "sql: no rows in result set" {
			return fmt.Errorf("404")
		}
		return errors.Wrap(err, "block: block get by hash failed")
	}
	return nil
}

// Insert new block in the database
func (b *Block) Insert() error {

	if err := b.pg.QueryRow(`
		INSERT INTO block
			(id, bits, height, nonce, version, hash_prev_block, hash_merkle_root, created_at, hash)
		VALUES
			(
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				$9
			)
			RETURNING id`,
		b.ID,
		b.Bits,
		b.Height,
		b.Nonce,
		b.Version,
		b.HashPrevBlock,
		b.HashMerkleRoot,
		b.CreatedAt,
		b.Hash,
	).Scan(&b.ID); err != nil {
		err = errors.Wrap(err, "block: insert block failed")
		return err
	}
	return nil
}

// getTransactions show transaction
func (b *Block) getTransactions() ([]Transaction, error) {
	var err error
	b.Transactions = make([]Transaction, 0, 0)

	if b.ID == 0 {
		return b.Transactions, nil
	}

	b.Transactions, err = FindTransactions(b.pg, "block_id", b.ID)
	if err != nil {
		return b.Transactions, errors.Wrap(err, "block: cannot get transactions")
	}

	return b.Transactions, nil
}

// getPrice return decimal price of bitcoin on the moment when block was created
func (b *Block) getPrice() (float32, error) {
	if err := b.pg.QueryRow(`SELECT 
		price
		FROM btc_price
		WHERE created_at >= $1 
		ORDER BY created_at ASC limit 1`, b.CreatedAt,
	).Scan(&b.Price); err != nil {
		return 0, errors.Wrap(err, "block: Cannot get bitcoin price")
	}
	return b.Price, nil
}

// NewBlock constructor for block structure
func NewBlock(pg *sql.DB) *Block {
	return &Block{
		pg: pg,
	}
}

// Last10 return last 10 blocks
func Last10(pg *sql.DB) []Block {
	blocks := make([]Block, 0)
	rows, err := pg.Query(`SELECT id, bits, height, nonce, version, hash_prev_block, hash_merkle_root, hash, created_at FROM block ORDER BY ID DESC LIMIT 10`)
	if err != nil {
		log.Fatalf("block: Cannot SELECT FROM BLOCK, %v", err)
	}

	for rows.Next() {
		var newB Block
		if err := rows.Scan(&newB.ID, &newB.Bits, &newB.Height, &newB.Nonce, &newB.Version, &newB.HashPrevBlock, &newB.HashMerkleRoot, &newB.Hash, &newB.CreatedAt); err != nil {
			log.Fatal(errors.Wrap(err, "block: Cannot retrieve block database data"))
		}
		newB.pg = pg
		blocks = append(blocks, newB)
	}
	return blocks
}
