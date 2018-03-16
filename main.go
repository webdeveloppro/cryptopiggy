package main

import (
	"log"
	"os"
)

var a App
var ws WsApp

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Please use webapp or wsapp parameter: ./bitcoin2sql <param>")
	}

	t := os.Args[1]

	if t == "webapp" {
		a.Initialize(
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USERNAME"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"))

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
