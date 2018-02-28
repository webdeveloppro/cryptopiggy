package main

import (
	"os"
)

var a App

func main() {
	a.Initialize(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	a.Run("")
}

/*
// GetInputAddress will return sender addres from signature script
func GetInputAddress(pubKeyHex string) (string, error) {
	decoded, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return "", err
	}
	// pack with sha256 once
	sha := chainhash.HashB(decoded)

	// encode with ripemd160 once
	rp := ripemd160.New()
	_, err = rp.Write(sha)
	if err != nil {
		return "", err
	}
	bcipher := rp.Sum(nil)

	// fill first byte with \x0
	one := make([]byte, 1)
	one[0] = 0x00
	bcipher = append(one[:], bcipher[:]...)

	// append data with last 4 bytes of sha256^2(data)
	res := append(bcipher[:], chainhash.DoubleHashB(bcipher)[:4]...)
	return base58.Encode(res), nil
}

func main() {
	db, err := database.Open("ffldb", os.Getenv("BTCD_DATADIR"), wire.MainNet)
	if err != nil {
		log.Fatalf("fatal error happenned: %v", err)
		// Handle error
	}

	connectionString :=
		fmt.Sprintf("host=%s user=%s password='%s' dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USERNAME"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
		)

	pg, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Cannot open postgresql connection: %v", err)
	}

	defer db.Close()
	defer pg.Close()

	cfg := blockchain.Config{
		DB:          db,
		ChainParams: &chaincfg.MainNetParams,
		TimeSource:  blockchain.NewMedianTime(),
	}

	bc, err := blockchain.New(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	blk, err := bc.BlockByHeight(2812)

	for _, tran := range blk.Transactions() {
		ins := tran.MsgTx().TxIn
		fmt.Println(tran.Hash())

		for _, txIn := range ins {
			// fmt.Printf("%d %s %d", i, txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
			if txIn.PreviousOutPoint.Hash.String() != "0000000000000000000000000000000000000000000000000000000000000000" {
				// spew.Dump(txIn)
				disbuf, err := txscript.DisasmString(txIn.SignatureScript)
				if err != nil {
					log.Fatalln(err)
				}
				disbufArr := strings.Split(disbuf, " ")
				// spew.Dump(disbuf, txIn.PreviousOutPoint)

				if len(disbufArr) > 1 {
					address, err := GetInputAddress(disbufArr[1])
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println("address: ", address)
				} else {
					prevTran, err := FindTransaction(pg, "hash", txIn.PreviousOutPoint.Hash.String())
					if err != nil {
						log.Fatal("Find tr error", err)
					}

					fmt.Println(prevTran.TxOuts[txIn.PreviousOutPoint.Index].Addresses)
				}
			}
		}
	}
}
*/
