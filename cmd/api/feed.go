package main

import (
	"net/http"

	"github.com/huynguyen1310/social/internal/store"
)

// getUserFeedHandler returns the feed for the current user
//
//	@Summary		Get user feed
//	@Description	Get a paginated feed of posts from users you follow
//	@Tags			feed
//	@Produce		json
//	@Param			limit	query		int		false	"Number of posts to return (1-20)"	default(20)
//	@Param			offset	query		int		false	"Number of posts to skip"			default(0)
//	@Param			sort	query		string	false	"Sort direction"					Enums(asc, desc)	default(desc)
//	@Param			tags	query		string	false	"Filter by tags (comma-separated)"	maxlength(100)
//	@Param			search	query		string	false	"Search in title or content"		maxlength(100)
//	@Param			since	query		string	false	"Return posts created after (RFC3339 format: 2006-01-02T15:04:05)"
//	@Param			until	query		string	false	"Return posts created before (RFC3339 format: 2006-01-02T15:04:05)"
//	@Success		200		{array}		store.PostWithMetadata
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/feeds [get]
func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fq := store.PaginationFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "asc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	feed, err := app.store.Posts.GetUserFeed(ctx, int64(13), fq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
