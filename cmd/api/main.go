package main

import (
	"log"
	"time"

	"github.com/huynguyen1310/social/internal/db"
	"github.com/huynguyen1310/social/internal/env"
	"github.com/huynguyen1310/social/internal/store"
	"go.uber.org/zap"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

//	@title			Social API
//	@version		1.0.0
//	@description	A social media API built with Go, Chi, and PostgreSQL.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found")
	}

	config := config{
		addr:   env.GetString("PORT", ":8081"),
		apiURL: env.GetString("API_URL", "localhost:8081"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost:5432/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 5),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		mail: mailConfig{
			exp: time.Hour * 24 * 3,
		},
	}

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db.New(
		config.db.addr,
		config.db.maxOpenConns,
		config.db.maxIdleConns,
		config.db.maxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}

	store := store.NewStorage(db)

	app := &application{
		config: config,
		store:  store,
		logger: logger,
	}

	defer db.Close()
	logger.Info("DB connect established")

	mux := app.mount()
	logger.Fatal(app.serve(mux))
}
