package main

import "net/http"

// healthHandler returns the service health status
//
//	@Summary		Health check
//	@Description	Check if the API is running
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func (app *application) healthHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status": "OK",
	}
	if err := writeJSON(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}
