package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx"
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
	DB     *pgx.Conn
}

// Initialize application and open db connection
func (a *WsApp) Initialize(pg *pgx.Conn) {

	a.DB = pg

	channel := "address_notify"
	err := a.DB.Listen(channel)
	if err != nil {
		log.Fatalf("Cannot subscribe on the %s channel, %v", channel, err)
	}

	channel = "blocks_notify"
	err = a.DB.Listen(channel)
	if err != nil {
		log.Fatalf("Cannot subscribe on the %s channel, %v", channel, err)
	}
}

// Run application on 8080 port
func (a *WsApp) Run(addr string) {

	if addr == "" {
		addr = "8082"
	}

	http.HandleFunc("/steam", a.addrsSteam)
	// http.HandleFunc("/steam_blocks", a.blocksSteam)
}

func (a *WsApp) addrsSteam(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	fmt.Println("block stream")
	tickChan := time.NewTicker(time.Second).C

	doneChan := make(chan bool)
	for {
		select {
		case <-tickChan:
			ctx := context.Background()
			notification, err := a.DB.WaitForNotification(ctx)
			if err != nil {
				log.Fatalf("cannot waitfornotification: %v", err)
				c.Close()
				// do something with notification
			}
			// in case if we will need to check channel manualy
			// res := fmt.Sprintf("%d%s", notification.Channel[0], notification.Payload)

			err = c.WriteMessage(1, []byte(notification.Payload))
			if err != nil {
				log.Println("write error:", err)
				c.Close()
				break
			}
		case <-doneChan:
			// defer c.Close()
			return
		}
	}
}

func (a *WsApp) blocksSteam(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	fmt.Println("block stream")
	tickChan := time.NewTicker(time.Second).C

	doneChan := make(chan bool)
	for {
		select {
		case <-tickChan:
			ctx := context.Background()
			notification, err := a.DB.WaitForNotification(ctx)
			if err != nil {
				log.Fatalf("cannot waitfornotification: %v", err)
				// do something with notification
			}
			err = c.WriteMessage(1, []byte(notification.Payload))
			if err != nil {
				log.Println("write error:", err)
				break
			}
		case <-doneChan:
			s := "done"
			err = c.WriteMessage(1, []byte(s))
			if err != nil {
				log.Println("write error:", err)
				break
			}
			return
		}
	}
	// defer c.Close()
}
