package block

import (
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/webdeveloppro/cryptopiggy/pkg/transaction"
)

// ErrNoPrice error for getPrice
var ErrNoPrice = fmt.Errorf("No price for coin")

// Storage is main interface for operations with Block
type Storage interface {
	GetByHash(string) (*Block, error)
	Insert(*Block) error
	Last10() ([]Block, error)
	getTransactions(uint) ([]transaction.Transaction, error)
	getPrice(time.Time) (float32, error)
}

// PGStorage provider that can handle read/write from database
type PGStorage struct {
	con *pgx.ConnPool
}

// NewStorage return storage reference
func NewStorage(pg *pgx.ConnPool) PGStorage {
	return PGStorage{
		con: pg,
	}
}

// GetByHash pull user from postgresql database
func (pg *PGStorage) GetByHash(hash string) (*Block, error) {

	bl := Block{storage: pg}
	if err := pg.con.QueryRow(`
		SELECT id, bits, height, nonce, version, hash_prev_block, hash_merkle_root, created_at, hash
		FROM block
		WHERE hash = $1
	`, hash).Scan(
		&bl.ID,
		&bl.Bits,
		&bl.Height,
		&bl.Nonce,
		&bl.Version,
		&bl.HashPrevBlock,
		&bl.HashMerkleRoot,
		&bl.CreatedAt,
		&bl.Hash,
	); err != nil {
		return &bl, err
	}
	return &bl, nil
}

// Insert new block in the database
func (pg *PGStorage) Insert(b *Block) error {

	// ToDo
	// Create each block in  pg commit/rollback transaction
	// BC2SQL.pg.BeginTx()
	if err := pg.con.QueryRow(`
		INSERT INTO block
			(bits, height, nonce, version, hash_prev_block, hash_merkle_root, created_at, hash)
		VALUES
			(
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8
			)
			RETURNING id`,
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

	for _, t := range b.Transactions {
		t.BlockID = b.ID
		if err := t.Insert(); err != nil {
			return err
		}
	}
	return nil
}

// Last10 gets last 10 transactions from database
func (pg *PGStorage) Last10() ([]Block, error) {
	blocks := make([]Block, 0)
	rows, err := pg.con.Query(`SELECT id, bits, height, nonce, version, hash_prev_block, hash_merkle_root, hash, created_at FROM block ORDER BY ID DESC LIMIT 10`)
	if err != nil {
		return blocks, errors.Wrapf(err, "block: Cannot SELECT FROM BLOCK, %v", err)
	}

	for rows.Next() {
		var newB Block
		if err := rows.Scan(&newB.ID, &newB.Bits, &newB.Height, &newB.Nonce, &newB.Version, &newB.HashPrevBlock, &newB.HashMerkleRoot, &newB.Hash, &newB.CreatedAt); err != nil {
			return blocks, errors.Wrap(err, "block: Cannot retrieve block database data")
		}
		blocks = append(blocks, newB)
	}
	return blocks, nil
}

// getTransactions show transaction
func (pg *PGStorage) getTransactions(id uint) ([]transaction.Transaction, error) {
	tranStorage := transaction.NewStorage(pg.con)
	return transaction.FindTransactions(tranStorage, "block_id", id)
}

// getPricePerBlock return decimal price of bitcoin on the moment when block was created
func (pg *PGStorage) getPrice(createdAt time.Time) (float32, error) {
	var price float32
	err := pg.con.QueryRow(`SELECT 
		price
		FROM btc_price
		WHERE created_at >= $1 
		ORDER BY created_at ASC limit 1`, createdAt,
	).Scan(&price)

	if err == pgx.ErrNoRows {
		return 0, ErrNoPrice
	}

	return price, err
}
