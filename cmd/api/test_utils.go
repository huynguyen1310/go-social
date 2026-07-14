package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/huynguyen1310/social/internal/auth"
	"github.com/huynguyen1310/social/internal/store"
	"github.com/huynguyen1310/social/internal/store/cache"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T) *application {
	t.Helper()
	logger := zap.Must(zap.NewProduction()).Sugar()
	mockStore := store.NewMockStorage(t)
	mockCacheStore := cache.NewMockStore()
	mockMailer := newMockMailer()
	testAuth := &auth.TestAuthenticator{}

	return &application{
		logger:        logger,
		store:         mockStore,
		cache:         mockCacheStore,
		mailer:        mockMailer,
		authenticator: testAuth,
		config: config{
			addr:   ":9999",
			apiURL: "localhost:9999",
			auth: authConfig{
				basic: basicAuthConfig{
					username: "admin",
					password: "password",
				},
				jwt: jwtAuthConfig{
					secret: "test-secret",
					aud:    "test-aud",
					iss:    "test-iss",
				},
			},
			cache: cacheConfig{
				enabled: false,
			},
		},
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, req)
	return recorder
}

type mockMailer struct{}

func newMockMailer() *mockMailer {
	return &mockMailer{}
}

func (m *mockMailer) Send(template string, username string, email string, data any) error {
	return nil
}
