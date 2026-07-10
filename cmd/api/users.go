package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/huynguyen1310/social/internal/store"
)

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := app.getUserFromCtx(r).ID

	user, err := app.store.Users.Get(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}

}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerID := app.getUserFromCtx(r).ID
	userID := int64(13)

	if err := app.store.Followers.Follow(r.Context(), followerID, userID); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			app.conflictError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
	}

}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerID := app.getUserFromCtx(r).ID
	userID := int64(13)

	if err := app.store.Followers.Unfollow(r.Context(), followerID, userID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
	}

}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		user, err := app.store.Users.Get(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, "user", user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (app *application) getUserFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value("user").(*store.User)
	return user
}
