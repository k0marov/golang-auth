package main

import (
	"auth"
	"log"
	"net/http"
)

const HashCost = 10
const dbFileName = "auth.db.json"

func main() {
	store, err := auth.NewStoreImpl(dbFileName)
	if err != nil {
		log.Fatal(err)
	}
	srv, err := auth.NewAuthServerImpl(store, HashCost)
	if err != nil {
		log.Fatal(err)
	}

	err = http.ListenAndServe(":4242", srv)
	if err != nil {
		log.Fatalf("could not listen on port 4242: %v", err)
	}
}
