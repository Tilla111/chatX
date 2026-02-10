package main

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error",
		"error", err,
		"request_id", middleware.GetReqID(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
	)

	writeJSONError(w, http.StatusInternalServerError, "internal server error")
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("bad request error",
		"error", err,
		"request_id", middleware.GetReqID(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
	)

	writeJSONError(w, http.StatusBadRequest, "bad request error")
}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("not found error",
		"error", err,
		"request_id", middleware.GetReqID(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
	)

	writeJSONError(w, http.StatusNotFound, "not found error")
}

func (app *application) ConflictError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("conflict error",
		"error", err,
		"request_id", middleware.GetReqID(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
	)

	writeJSONError(w, http.StatusConflict, "Conflict error Cuncurrent")
}
