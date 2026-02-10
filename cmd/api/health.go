package main

import (
	"net/http"
)

func (app *application) healthCheck(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "available",
		"version": version,
		"message": "Welcome to ChatX API",
		"ENV":     app.config.ENV,
	}
	err := writeJSON(w, http.StatusOK, data)
	if err != nil {
		app.InternalServerError(w, r, err)
	}
}
