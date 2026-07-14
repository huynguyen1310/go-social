package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("internal server error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())

	writeJSONError(w, http.StatusInternalServerError, "the server encounted a problem")
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("bad request error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("conflict error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())

	writeJSONError(w, http.StatusConflict, err.Error())
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("not found error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())

	writeJSONError(w, http.StatusNotFound, err.Error())
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("bad request error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) unauthorizedBasicError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("unauthorized basic error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	writeJSONError(w, http.StatusUnauthorized, err.Error())
}

func (app *application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("unauthorized error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusUnauthorized, err.Error())
}

func (app *application) forbiddenErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("forbidden error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	writeJSONError(w, http.StatusForbidden, err.Error())
}

func (app *application) rateLimitExceededErrorResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Warnf("rate limit exceeded error: %s path: %s error: %s", r.Method, r.URL.Path)

	w.Header().Set("Retry-After", retryAfter)

	writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after "+retryAfter)
}
