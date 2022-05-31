package main

import (
	"auth/core/crypto/bcrypt_hasher"
	"auth/data/store"
	"auth/data/store/db_file_interactor_impl"
	"auth/delivery/server"
	"auth/domain/auth_service"
	"log"
	"net/http"
	"os"
)

const HashCost = 10
const dbFileName = "auth.db.json"

func main() {
	dbFile, err := os.OpenFile(dbFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("problem opening %s %v", dbFileName, err)
	}
	fileInteractor := db_file_interactor_impl.NewDBFileInteractor(dbFile)
	store, err := store.NewPersistentInMemoryFileStore(fileInteractor)
	if err != nil {
		log.Fatalf("problem creating a store: %v", err)
	}

	hasher := bcrypt_hasher.NewBcryptHasher(HashCost)
	service := auth_service.NewAuthServiceImpl(store, hasher)
	srv := server.NewAuthServer(service)

	err = http.ListenAndServe(":4242", srv)
	if err != nil {
		log.Fatalf("could not listen on port 4242: %v", err)
	}
}
