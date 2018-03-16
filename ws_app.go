package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var upgrader = websocket.Upgrader{
	// allow connection from any place
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WsApp holding hash channel btcd client and db connections
type WsApp struct {
	hash   chan string
	client *rpcclient.Client
	DB     *sql.DB
}

// Initialize application and open db connection
func (a *WsApp) Initialize(host, user, password, dbname string) {

	if host == "" {
		log.Fatal("Empty host string, setup DB_HOST env")
		host = "localhost"
	}

	if user == "" {
		log.Fatal("Empty user string, setup DB_USER env")
		return
	}

	if dbname == "" {
		log.Fatal("Empty dbname string, setup DB_DBNAME env")
		return
	}

	connectionString :=
		fmt.Sprintf("host=%s user=%s password='%s' dbname=%s sslmode=disable", host, user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Cannot open postgresql connection: %v", err)
	}
	//defer a.DB.Close()
	a.hash = make(chan string)

	// Connect to local btcd RPC server using websockets.
	btcdHomeDir := btcutil.AppDataDir("btcd", false)
	certs, err := ioutil.ReadFile(filepath.Join(btcdHomeDir, "rpc.cert"))
	if err != nil {
		log.Fatal(err)
	}
	connCfg := &rpcclient.ConnConfig{
		Host:         "127.0.0.1:8334",
		Endpoint:     "ws",
		User:         "admin",
		Pass:         "123123123",
		Certificates: certs,
	}

	// Only override the handlers for notifications you care about.
	// Also note most of these handlers will only be called if you register
	// for notifications.  See the documentation of the rpcclient
	// NotificationHandlers type for more details about each handler.

	ntfnHandlers := rpcclient.NotificationHandlers{
		OnFilteredBlockConnected: a.NewBlockNotification,
	}

	a.client, err = rpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		log.Fatal(err)
	}

	// Register for block connect and disconnect notifications.
	if err := a.client.NotifyBlocks(); err != nil {
		log.Fatal(err)
	}
}

// NewBlockNotification send new blockchain hash to channel
func (a *WsApp) NewBlockNotification(height int32, header *wire.BlockHeader, txns []*btcutil.Tx) {
	hash := header.BlockHash()
	log.Printf("Block connected: %v (%d) %v\n",
		hash, height, header.Timestamp)

	go func() {
		a.hash <- hash.String()
	}()
	fmt.Printf("put to the channel: %s\n", hash.String())
}

// Run application on 8080 port
func (a *WsApp) Run(addr string) {

	if addr == "" {
		addr = "8082"
	}

	http.HandleFunc("/steam", a.blockSteam)
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func (a *WsApp) blockSteam(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	// defer c.Close()
	go func() {
		for {
			select {
			case s := <-a.hash:
				// do stuff
				hash, err := chainhash.NewHashFromStr(s)
				if err != nil {
					log.Fatalf("Cannot get hash from string, %+v", err)
				}

				wr, err := a.client.GetBlock(hash)
				if err != nil {
					log.Fatal(err)
				}
				type ret struct {
					ID               int32
					Hash             string
					Hash_merkle_root string
					Prev_hash_block  string
				}
				r := &ret{
					ID:               1,
					Hash:             hash.String(),
					Hash_merkle_root: wr.Header.MerkleRoot.String(),
					Prev_hash_block:  wr.Header.PrevBlock.String(),
				}

				fmt.Printf("header: %+v\n", r)

				buf, err := json.Marshal(r)
				if err != nil {
					log.Fatalf("cannot marshal header %v", err)
				}
				/*
					var buf bytes.Buffer
					if err := binary.Write(&buf, binary.LittleEndian, wr.Header); err != nil {
						log.Fatalf("Cannot write header to binary, %v", err)
					}
				*/
				err = c.WriteMessage(1, buf)
				if err != nil {
					log.Println("write error:", err)
					break
				}
			}
		}
	}()
}
