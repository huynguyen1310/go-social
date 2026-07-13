package main

import (
	"log"

	"github.com/huynguyen1310/social/internal/db"
	"github.com/huynguyen1310/social/internal/env"
	"github.com/huynguyen1310/social/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost:5432/social?sslmode=disable")
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	store := store.NewStorage(conn)
	if err := db.Seed(store, conn); err != nil {
		log.Fatal(err)
	}
}
