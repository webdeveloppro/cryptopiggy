package block

import (
	"testing"

	"github.com/jackc/pgx"
	"github.com/vladyslav2/bitcoin2sql/pkg/transaction"
)

type FakeStorage struct {
	resp string
	code int
}

func (s FakeStorage) GetByHash(hash string) (*Block, error) {
	bl := Block{
		ID:             154728,
		Bits:           437129626,
		Height:         154724,
		Nonce:          3419332106,
		Version:        1,
		HashPrevBlock:  "00000000000003bfc715be0afb06486c325c12dea913766564fd7e9bc453889d",
		HashMerkleRoot: "4eff3116a1a55119f83e829f30761692695837e28abe2d4d2ba1b03aecdbd0d9",
		CreatedAt:      "2011-11-25 04:52:48",
		Hash:           "0000000000000aece46da94d3880c3d43c3da17a1e7f06f5ca199aad9dbbac3e",
	}

	if hash == "existhash" {
		return &bl, nil
	}

	if hash == "notexisthash" {
		return &bl, pgx.ErrNoRows
	}

	panic("hash do not match anything, please verify email address")
}

func (s FakeStorage) Insert(*Block) error {
	return nil
}

func (s FakeStorage) Last10() ([]Block, error) {
	return make([]Block, 5), nil
}

func (s FakeStorage) getTransactions(int) ([]transaction.Transaction, error) {
	return make([]transaction.Transaction, 15), nil
}

func (s FakeStorage) getPrice(CreatedAt string) (float32, error) {
	if CreatedAt == "right_date" {
		return 100.00, nil
	}
	return 0, ErrNoPrice
}

func TestGetByHash(t *testing.T) {
	t.Parallel()
	s := FakeStorage{}

	b, err := s.GetByHash("existhash")
	if b.ID == 0 {
		t.Errorf("GetByHash cant find a block but should, block: %v, err: %v", b, err)
	}

	b2, err := s.GetByHash("notexisthash")
	if err != pgx.ErrNoRows {
		t.Errorf("GetByHash return wrong error for not existing block, %v, %v", b2, err)
	}
}

func TestLast10(t *testing.T) {
	t.Parallel()

	trans, _ := Last10(FakeStorage{})
	if len(trans) != 5 {
		t.Errorf("Last10 return wrong amount should 5, got: %v", trans)
	}
}

func TestGetTransactions(t *testing.T) {
	t.Parallel()
	b := New(FakeStorage{})

	trans, _ := b.GetTransactions()
	if len(trans) != 0 {
		t.Errorf("getTransactions return wrong amount should 0, got: %v", trans)
	}

	b.ID = 123
	trans, _ = b.GetTransactions()
	if len(trans) != 15 {
		t.Errorf("getTransactions return wrong amount should 15, got: %v", trans)
	}
}

func TestGetPrice(t *testing.T) {
	t.Parallel()
	b := New(FakeStorage{})
	b.CreatedAt = "right_date"

	err := b.GetPrice()
	if b.Price != 100.00 {
		t.Errorf("getPrice return wrong amount should 100.00, got: %v", b.Price)
	}

	b.CreatedAt = "not_right_date"
	err = b.GetPrice()
	if err != ErrNoPrice {
		t.Errorf("getPrice return wrong err message, should: %v, got: %v", ErrNoPrice, err)
	}
}
