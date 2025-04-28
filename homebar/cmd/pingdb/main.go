package main

import (
	"log"

	"github.com/kexincchen/homebar/internal/config"
	"github.com/kexincchen/homebar/internal/db"
)

func main() {
	cfg := config.Load()

	if _, err := db.NewPostgres(cfg); err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	log.Println("Postgres connection OK")
}
