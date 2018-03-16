package transaction

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/pkg/errors"
)

type FakeStorage struct {
	resp string
	code int
}

func readJSONFile(path string, out interface{}) error {

	txBuf, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "Cannot read %s", path)
	}

	if err := json.Unmarshal(txBuf, &out); err != nil {
		return errors.Wrapf(err, "Cannot parse txBuf %s to json, %s", path, txBuf)
	}
	return nil
}

func (s FakeStorage) GetByWhere(sql string, val ...interface{}) ([]Transaction, error) {
	trans := make([]Transaction, 0)

	if val[0].(string) == "good_wallet" {
		// Txin + Txouts
		txin1 := make([]TxIn, 0)
		err := readJSONFile("fixtures/txin_1.json", &txin1)
		if err != nil {
			return trans, err
		}

		txout1 := make([]TxOut, 0)
		err = readJSONFile("fixtures/txout_1.json", &txout1)
		if err != nil {
			return trans, err
		}

		txin2 := make([]TxIn, 0)
		err = readJSONFile("fixtures/txin_2.json", &txin2)
		if err != nil {
			return trans, err
		}

		txout2 := make([]TxOut, 0)
		err = readJSONFile("fixtures/txout_2.json", &txout2)
		if err != nil {
			return trans, err
		}

		return []Transaction{
			Transaction{
				ID:         1915892,
				BlockID:    154734,
				Hash:       "669c479e1eb2bfcc25cf1406c9b6b922cf7dea5691d5e85d4cbcc4b32464f93d",
				HasWitness: false,
				TxIns:      txin1,
				TxOuts:     txout1,
			},
			Transaction{
				ID:         1915936,
				BlockID:    154735,
				Hash:       "301f4d8f32248727810b25bcac8f561455dc6821ff18a65a03b6f834ab475000",
				HasWitness: false,
				TxIns:      txin2,
				TxOuts:     txout2,
			},
		}, nil
	}

	if val[0].(string) == "bad_wallet" {
		return trans, nil
	}

	panic("hash do not match anything, please verify val data")
}

func (s FakeStorage) GetPricePerTransaction(trans []Transaction) error {
	for i := range trans {
		trans[i].Price = 75.00
	}
	return nil
}

func TestFindTransactions(t *testing.T) {
	t.Parallel()
	f := FakeStorage{}

	trans, err := FindTransactions(f, "address_hash", "good_wallet")
	if len(trans) != 2 {
		t.Errorf("FindTransactions return wrong results for address_hash: %v, %v", trans, err)
	}

	trans, err = FindTransactions(f, "address_hash", "bad_wallet")
	if len(trans) != 0 {
		t.Errorf("FindTransactions return wrong results for address_hash: %v, %v", trans, err)
	}
}

func TestFindTransaction(t *testing.T) {
	t.Parallel()

	f := FakeStorage{}
	trans, err := FindTransaction(f, "address_hash", "good_wallet")
	if trans.Hash != "669c479e1eb2bfcc25cf1406c9b6b922cf7dea5691d5e85d4cbcc4b32464f93d" {
		t.Errorf("FindTransactions return wrong results for goodval address, %v", trans)
	}

	trans, err = FindTransaction(f, "address_hash", "bad_wallet")
	if err != ErrNoTran {
		t.Errorf("FindTransactions return wrong results for badval, %v, trans: %v", err, trans)
	}
}

func TestGetAddresses(t *testing.T) {
	t.Parallel()

	txout1 := make([]TxOut, 0)
	if err := readJSONFile("fixtures/txout_1.json", &txout1); err != nil {
		t.Errorf("Cannot read json txout_1.json, %v", err)
	}

	addrs, err := txout1[0].getAddresses()

	if err != nil {
		t.Error(err)
	}

	if len(addrs) == 0 {
		t.Errorf("txout should return addresses but return not result, %+v", txout1[0])
	}

	if addrs[0] != "1LPXQf1foebcfLzZxcpK3sG2TJ9ke1uLyQ" {
		t.Errorf("Got wrong address expected 1LPXQf1foebcfLzZxcpK3sG2TJ9ke1uLyQ got: %s", addrs[0])
	}

	if err := readJSONFile("fixtures/txout_2.json", &txout1); err != nil {
		t.Errorf("Cannot read json txout_2.json, %v", err)
	}

	addrs, err = txout1[1].getAddresses()

	if err != nil {
		t.Error(err)
	}

	if addrs[0] != "1AUzdKPJtPyFgrg88nK6fBH7jKkgPyphUj" {
		t.Errorf("Got wrong address expected 1AUzdKPJtPyFgrg88nK6fBH7jKkgPyphUj got: %s", addrs[0])
	}
}

func TestGetPricePerTransaction(t *testing.T) {
	t.Parallel()
	f := FakeStorage{}
	trans := make([]Transaction, 3)

	GetPricePerTransaction(f, trans)

	for _, f := range trans {
		if f.Price != 75.00 {
			t.Errorf("TestGetPricePerTransaction return wrong amount should 75.00, got: %f", f.Price)
		}
	}
}
