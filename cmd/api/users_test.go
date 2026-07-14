package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGetUser(t *testing.T) {
	app := newTestApplication(t)
	mux := app.mount()
	claims := jwt.MapClaims{
		"sub": "1",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"iss": "test-iss",
		"aud": "test-aud",
	}
	testToken, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("should not allow unauthenticated request", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := executeRequest(req, mux)
		checkResponseCode(t, rr.Code, http.StatusUnauthorized)

	})

	t.Run("should allow authenticated request", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)
		checkResponseCode(t, rr.Code, http.StatusOK)
	})

}

func checkResponseCode(t *testing.T, code, expectedCode int) {
	if code != expectedCode {
		t.Errorf("expected status code %d, got %d", expectedCode, code)
	}
}
