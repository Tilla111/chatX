package main

import (
	"net/http"
)

// healthCheck godoc
// @Summary      API holatini tekshirish
// @Description  API ishlayotganini tekshirish uchun texnik endpoint.
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]string  "{"status":"available","version":"v1.0.0","message":"Welcome to ChatX API","ENV":"dev"}"
// @Failure      500  {object}  map[string]string  "Ichki server xatosi"
// @Router       /health [get]
func (app *application) healthCheck(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "available",
		"version": Version,
		"message": "Welcome to ChatX API",
		"ENV":     app.config.ENV,
	}
	err := writeJSON(w, http.StatusOK, data)
	if err != nil {
		app.internalServerError(w, r, err)
	}
}
