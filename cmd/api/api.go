package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/huynguyen1310/social/docs"
	"github.com/huynguyen1310/social/internal/auth"
	"github.com/huynguyen1310/social/internal/mailer"
	"github.com/huynguyen1310/social/internal/store"
	"github.com/huynguyen1310/social/internal/store/cache"
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
}

type config struct {
	addr   string
	db     dbConfig
	apiURL string
	mail   mailConfig
	auth   authConfig
	cache  cacheConfig
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

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

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

	app.logger.Infof("listening on %s", app.config.addr)

	return server.ListenAndServe()
}
