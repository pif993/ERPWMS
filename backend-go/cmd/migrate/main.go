package main

import (
	"log"
	"os"

	"github.com/pressly/goose/v3"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL required")
	}
	db, err := goose.OpenDBWithDriver("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := goose.Up(db, "internal/db/migrations"); err != nil {
		log.Fatal(err)
	}
}
