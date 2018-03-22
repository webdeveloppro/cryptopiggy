package transaction

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/pkg/errors"
)

// ErrNonStandard Error for non standart output address
var ErrNonStandard = fmt.Errorf("Non standard output")

// TxIn transaction incoming data
type TxIn struct {
	Amount          int64  `json:"amount"`
	PrevOut         string `json:"prev_out"`
	Size            int    `json:"size"`
	SignatureScript string `json:"signature_script"`
	Sequence        uint32 `json:"sequence"`
	Witness         string `json:"witness"`
	Address         string `json:"address"`
	AddressID       uint   `json:"address_id"`
}

// TxOut transaction outcoming data
type TxOut struct {
	PkScript  string   `json:"pk_script"` // Hex version of PkScript
	Value     int64    `json:"val"`
	Addresses []string `json:"addresses"`
}

// GetAddresses return addresses where money went
func (txOut *TxOut) GetAddresses() ([]string, error) {
	if txOut.Addresses == nil {
		dst := make([]byte, hex.DecodedLen(len(txOut.PkScript)))
		_, err := hex.Decode(dst, []byte(txOut.PkScript))
		if err != nil {
			return []string{}, errors.Wrap(err, "block: Cannot convert hex string to bytes")
		}

		typ, addresses, _, err := txscript.ExtractPkScriptAddrs(dst, &chaincfg.MainNetParams)
		if err != nil {
			return []string{}, errors.Wrap(err, fmt.Sprintf("Cannot extract pkScript %s", txOut.PkScript))
		}

		addrs := []string{}

		// dirty hack for weired transactions
		// like 9969603dca74d14d29d1d5f56b94c7872551607f8c2d6837ab9715c60721b50e
		if typ == txscript.NonStandardTy {
			// We still need to create some Addresses to work correctly, so we just create any hash
			guessedAddr, err := txOut.guessNonstandardHash()
			if err != nil {
				return nil, err
			}
			addrs = append(addrs, guessedAddr)
		} else {
			for _, a := range addresses {
				addrs = append(addrs, a.EncodeAddress())
			}
		}
		txOut.Addresses = addrs
	}
	return txOut.Addresses, nil
}

// For Nonstandard output transactions we still need to get some hash
// it is fix for being able to work with badly/unclear transactions
// we will generate some hash from pkscript data
func (txOut *TxOut) guessNonstandardHash() (string, error) {

	log.Printf("Nonstandard txout: %s", txOut.PkScript)

	r := txOut.PkScript
	if len(r) > 22 {
		r = r[0:22]
	}

	return fmt.Sprintf("nonstandard-%s", r), nil
}
