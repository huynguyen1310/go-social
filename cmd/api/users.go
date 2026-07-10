package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/huynguyen1310/social/internal/store"
)

// getUserHandler returns a user by ID
//
//	@Summary		Get user by ID
//	@Description	Get a specific user by their ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int	true	"User ID"
//	@Success		200		{object}	store.User
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Router			/users/{userID} [get]
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

// followUserHandler follows a user
//
//	@Summary		Follow a user
//	@Description	Follow another user by their ID
//	@Tags			users
//	@Produce		json
//	@Param			userID	path		int	true	"User ID to follow"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		409		{object}	error
//	@Failure		500		{object}	error
//	@Router			/users/{userID}/follow [put]
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

// unfollowUserHandler unfollows a user
//
//	@Summary		Unfollow a user
//	@Description	Unfollow another user by their ID
//	@Tags			users
//	@Produce		json
//	@Param			userID	path		int	true	"User ID to unfollow"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Router			/users/{userID}/unfollow [put]
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
