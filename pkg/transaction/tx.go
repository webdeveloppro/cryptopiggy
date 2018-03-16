package transaction

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/pkg/errors"
)

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
	Value     int64    `json:"val"`
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
