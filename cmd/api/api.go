package main

import (
	"context"
	"database/sql"
	"errors"
	"expvar"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/huynguyen1310/social/docs"
	"github.com/huynguyen1310/social/internal/auth"
	"github.com/huynguyen1310/social/internal/mailer"
	ratelimiter "github.com/huynguyen1310/social/internal/rateLimiter"
	"github.com/huynguyen1310/social/internal/store"
	"github.com/huynguyen1310/social/internal/store/cache"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         store.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	cache         cache.Store
	rateLimiter   ratelimiter.Limiter
	db            *sql.DB
	rdb           *redis.Client
}

type config struct {
	addr        string
	db          dbConfig
	apiURL      string
	mail        mailConfig
	auth        authConfig
	cache       cacheConfig
	rateLimiter ratelimiter.Config
}

type cacheConfig struct {
	addr     string
	password string
	db       int
	enabled  bool
}

type authConfig struct {
	basic basicAuthConfig
	jwt   jwtAuthConfig
}

type basicAuthConfig struct {
	username string
	password string
}

type jwtAuthConfig struct {
	secret string
	aud    string
	iss    string
}

type mailConfig struct {
	exp       time.Duration
	fromEmail string
	smtpHost  string
	smtpPort  int
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(app.rateLimiterMiddleware)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// Publish DB pool stats
	expvar.Publish("db", expvar.Func(func() any {
		stats := app.db.Stats()
		return map[string]any{
			"max_open_connections": stats.MaxOpenConnections,
			"open_connections":     stats.OpenConnections,
			"in_use":               stats.InUse,
			"idle":                 stats.Idle,
			"wait_count":           stats.WaitCount,
			"wait_duration_ms":     stats.WaitDuration.Milliseconds(),
			"max_idle_closed":      stats.MaxIdleClosed,
			"max_idle_time_closed": stats.MaxIdleTimeClosed,
			"max_lifetime_closed":  stats.MaxLifetimeClosed,
		}
	}))

	// Publish Redis pool stats
	if app.rdb != nil {
		expvar.Publish("redis", expvar.Func(func() any {
			stats := app.rdb.PoolStats()
			return map[string]any{
				"hits":        stats.Hits,
				"misses":      stats.Misses,
				"timeouts":    stats.Timeouts,
				"total_conns": stats.TotalConns,
				"idle_conns":  stats.IdleConns,
			}
		}))
	}

	r.Get("/debug/vars", expvar.Handler().ServeHTTP)

	r.Route("/v1", func(r chi.Router) {
		r.With(app.BasicAuthMiddleware).Get("/health", app.healthHandler)

		r.Get("/swagger/*", httpSwagger.Handler())
		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			r.Post("/", app.createPostHandler)

			r.Route("/{postID}", func(r chi.Router) {
				r.Use(app.postContextMiddleware)

				r.Get("/", app.getPostHandler)
				r.Delete("/", app.checkPostOwnership("moderator", app.deletePostHandler))
				r.Patch("/", app.checkPostOwnership("admin", app.updatePostHandler))
			})
		})

		r.Route("/users", func(r chi.Router) {
			// /feeds must be registered BEFORE /{userID} so chi doesn't
			// match "feeds" as a userID parameter and fail to parse it as int64.
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feeds", app.getUserFeedHandler)
			})

			r.Get("/activate/{token}", app.activateUserHandler)

			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Use(app.userContextMiddleware)

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})

	return r

}

func (app *application) serve(mux http.Handler) error {
	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	server := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	shutdown := make(chan error, 1)
	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Infow("signal caught", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shutdown <- server.Shutdown(ctx)
	}()

	app.logger.Infof("listening on %s", app.config.addr)

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdown; err != nil {
		return err
	}

	// Cleanup resources
	if app.db != nil {
		if err := app.db.Close(); err != nil {
			app.logger.Errorf("db close failed: %v", err)
		}
	}
	if app.rdb != nil {
		if err := app.rdb.Close(); err != nil {
			app.logger.Errorf("redis close failed: %v", err)
		}
	}

	app.logger.Info("server stopped gracefully")
	return nil
}
