package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/huynguyen1310/social/internal/store"
)

func (app *application) BasicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedBasicError(w, r, fmt.Errorf("authorization header is missing"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Basic" {
			app.unauthorizedBasicError(w, r, fmt.Errorf("invalid authorization header"))
			return
		}

		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			app.unauthorizedBasicError(w, r, err)
			return
		}

		username := app.config.auth.basic.username
		password := app.config.auth.basic.password

		credentials := strings.SplitN(string(decoded), ":", 2)
		if len(credentials) != 2 || credentials[0] != username || credentials[1] != password {
			app.unauthorizedBasicError(w, r, fmt.Errorf("invalid credentials"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("invalid authorization header"))
			return
		}

		token := parts[1]
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("invalid token"))
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(claims["sub"].(string), 10, 64)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("invalid token"))
			return
		}

		ctx := r.Context()
		user, err := app.getUser(ctx, userID)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("invalid token"))
			return
		}

		ctx = context.WithValue(ctx, "auth_user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) checkPostOwnership(role string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getAuthUserFromCtx(r)
		post := app.getPostFormCtx(r)

		// Post owner can always proceed
		if post.UserId == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		// Not the owner — check role precedence
		allowed, err := app.checkRolePrecedence(r.Context(), role, user)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, fmt.Errorf("role check failed"))
			return
		}
		if !allowed {
			app.forbiddenErrorResponse(w, r, fmt.Errorf("you do not have the required role"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, roleName string, user *store.User) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil

}

func (app *application) getUser(ctx context.Context, userId int64) (*store.User, error) {
	if app.config.cache.enabled {
		user, err := app.cache.Users.Get(ctx, userId)
		if err != nil {
			return nil, err
		}

		if user != nil {
			app.logger.Infow("cache hit", "key", "user", "id", userId)
			return user, nil
		}

		app.logger.Infow("cache miss", "key", "user", "id", userId)
	}

	user, err := app.store.Users.Get(ctx, userId)
	if err != nil {
		return nil, err
	}

	if app.config.cache.enabled {
		if err := app.cache.Users.Set(ctx, user); err != nil {
			return nil, err
		}
	}

	return user, nil
}
