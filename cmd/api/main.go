package main

import (
	"log"
	"os"

	"github.com/matheuszzy/ledger/internal/account/handler"
	"github.com/matheuszzy/ledger/internal/account/repository"
	"github.com/matheuszzy/ledger/internal/account/service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := handler.New(service.New(repository.NewMemory())).Router()

	log.Fatal(router.Run(":" + port))
}
