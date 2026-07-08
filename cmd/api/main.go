package main

import (
	"log"

	"github.com/huynguyen1310/social/internal/db"
	"github.com/huynguyen1310/social/internal/env"
	"github.com/huynguyen1310/social/internal/store"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found")
	}

	config := config{
		addr: env.GetString("PORT", ":8081"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost:5432/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 5),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
	}

	db, err := db.New(
		config.db.addr,
		config.db.maxOpenConns,
		config.db.maxIdleConns,
		config.db.maxIdleTime,
	)
	if err != nil {
		log.Fatal(err)
	}

	store := store.NewStorage(db)

	app := &application{
		config: config,
		store:  store,
	}

	defer db.Close()
	log.Println("DB connect established")

	mux := app.mount()
	log.Fatal(app.serve(mux))
}
