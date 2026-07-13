package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/huynguyen1310/social/internal/mailer"
	"github.com/huynguyen1310/social/internal/store"
)

type RegisterRequest struct {
	Username string `json:"username" validate:"required,max=255"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

// registerUserHandler creates a new user account
//
//	@Summary		Register a new user
//	@Description	Create a new user account with username, email, and password. An invitation token is generated for email verification.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			user	body		RegisterRequest	true	"User registration details"
//
//	@Success		201		{object}	UserWithToken
//
//	@Failure		400		{object}	error
//	@Failure		409		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/register [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterRequest

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	user := store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashedToken := hex.EncodeToString(hash[:])

	ctx := r.Context()
	if err := app.store.Users.CreateAndInvite(ctx, &user, hashedToken, app.config.mail.exp); err != nil {
		switch err {
		case store.ErrDuplicateEmail, store.ErrDuplicateUsername:
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	// Send activation email asynchronously
	activationLink := fmt.Sprintf("http://%s/v1/users/activate/%s", app.config.apiURL, plainToken)
	go func() {
		data := mailer.ActivationData{
			Username:       user.Username,
			ActivationLink: activationLink,
		}
		if err := app.mailer.Send(mailer.ActivationTemplate, user.Username, user.Email, data); err != nil {
			app.logger.Errorw("failed to send activation email", "error", err, "email", user.Email)
		}
	}()

	userWithToken := UserWithToken{
		User:  &user,
		Token: plainToken,
	}

	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
	}
}

// activateUserHandler activates a user account by invitation token
//
//	@Summary		Activate user account
//	@Description	Activate a user account using the invitation token received via email
//	@Tags			auth
//	@Produce		html
//	@Param			token	path	string	true	"Invitation token"
//	@Success		200		"HTML page"
//	@Failure		404		"HTML error page"
//	@Failure		500		"HTML error page"
//
//	@Router			/users/activate/{token} [get]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.renderHTML(w, http.StatusNotFound, "Invalid or expired activation link.")
		default:
			app.renderHTML(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
		}
		return
	}

	app.renderHTML(w, http.StatusOK, "Your account has been activated! You can now log in.")
}

type CreateTokenRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// createTokenHandler authenticates a user and returns a JWT token
//
//	@Summary		Create a new JWT token
//	@Description	Authenticate with email and password, returns a signed JWT token for subsequent requests
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		CreateTokenRequest	true	"Login credentials"
//	@Success		201			{object}	map[string]any
//	@Failure		400			{object}	error
//	@Failure		401			{object}	error
//	@Failure		500			{object}	error
//	@Router			/auth/token [post]
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateTokenRequest
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.unauthorizedErrorResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	claims := jwt.MapClaims{
		"sub": strconv.FormatInt(user.ID, 10),
		"exp": time.Now().Add(time.Minute * 15).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.jwt.iss,
		"aud": app.config.auth.jwt.aud,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
	}

}
