package main

import "net/http"

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	feed, err := app.store.Posts.GetUserFeed(ctx, int64(13))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
