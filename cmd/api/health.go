package main

import (
	"net/http"
)

// healthCheck godoc
// @Summary      API holatini tekshirish
// @Description  Serverning ishchi holati va versiyasini qaytaradi.
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]map[string]string "Masalan: {"data": {"status": "available", "version": "1.0.0"}}"
// @Failure      500  {object}  map[string]string
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
		app.InternalServerError(w, r, err)
	}
}
