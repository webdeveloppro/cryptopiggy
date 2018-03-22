package main

import (
	"log"
	"os"

	"github.com/jackc/pgx"
)

var a App
var ws WsApp

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Please use webapp or wsapp parameter: ./bitcoin2sql <param>")
	}

	t := os.Args[1]

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USERNAME")
	dbpassword := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" {
		log.Print("Empty host string, setup DB_HOST env")
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

	connPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     host,
			User:     user,
			Password: dbpassword,
			Database: dbname,
		},
		MaxConnections: 100,
	}

	pool, err := pgx.NewConnPool(connPoolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool %v", err)
	}

	if t == "webapp" {
		a.Initialize(pool)
		a.Run("")
	} else if t == "wsapp" {
		ws.Initialize(
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USERNAME"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"))

		ws.Run("")
	}
}
