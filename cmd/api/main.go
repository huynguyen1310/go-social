package main

import (
	"log"
	"time"

	"github.com/huynguyen1310/social/internal/auth"
	"github.com/huynguyen1310/social/internal/db"
	"github.com/huynguyen1310/social/internal/env"
	"github.com/huynguyen1310/social/internal/mailer"
	ratelimiter "github.com/huynguyen1310/social/internal/rateLimiter"
	"github.com/huynguyen1310/social/internal/store"
	"github.com/huynguyen1310/social/internal/store/cache"
	"github.com/redis/go-redis/v9"
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
			exp:       time.Hour * 24 * 3,
			fromEmail: env.GetString("MAIL_FROM_EMAIL", "social@example.com"),
			smtpHost:  env.GetString("MAIL_SMTP_HOST", "localhost"),
			smtpPort:  env.GetInt("MAIL_SMTP_PORT", 1025),
		},
		auth: authConfig{
			basic: basicAuthConfig{
				username: env.GetString("AUTH_BASIC_USERNAME", ""),
				password: env.GetString("AUTH_BASIC_PASSWORD", ""),
			},
			jwt: jwtAuthConfig{
				secret: env.GetString("AUTH_JWT_SECRET", ""),
				aud:    env.GetString("AUTH_JWT_AUD", ""),
				iss:    env.GetString("AUTH_JWT_ISS", ""),
			},
		},
		cache: cacheConfig{
			addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
			password: env.GetString("REDIS_PASSWORD", ""),
			db:       env.GetInt("REDIS_DB", 0),
			enabled:  env.GetBool("REDIS_ENABLED", true),
		},
		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.GetInt("RATE_LIMIT", 100),
			TimeFrame:            env.GetDuration("RATE_LIMIT_WINDOW", time.Minute),
			Enabled:              env.GetBool("RATE_LIMIT_ENABLED", true),
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

	mailer := mailer.NewMailer(config.mail.smtpHost, config.mail.smtpPort, config.mail.fromEmail)
	if mailer == nil {
		logger.Fatal("failed to create mailer")
	}

	jwtAuthenticator := auth.NewJWTAuthenticator(
		config.auth.jwt.secret,
		config.auth.jwt.aud,
		config.auth.jwt.iss,
	)
	if jwtAuthenticator == nil {
		logger.Fatal("failed to create JWT authenticator")
	}

	var rdb *redis.Client
	if config.cache.enabled {
		rdb = cache.NewRedisClient(
			config.cache.addr,
			config.cache.password,
			config.cache.db,
		)
		logger.Info("cache connect established")
	}

	cacheStore := cache.NewRedisStore(rdb)

	rateLimiter := ratelimiter.NewFixedWindowRateLimiter(
		config.rateLimiter.RequestsPerTimeFrame,
		config.rateLimiter.TimeFrame,
	)

	app := &application{
		config:        config,
		store:         store,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
		cache:         *cacheStore,
		rateLimiter:   rateLimiter,
		db:            db,
		rdb:           rdb,
	}

	logger.Info("DB connect established")

	mux := app.mount()

	if err := app.serve(mux); err != nil {
		logger.Fatal(err)
	}
}
