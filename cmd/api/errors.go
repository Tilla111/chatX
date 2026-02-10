package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
)

func (app *application) InternalServerError(w http.ResponseWriter, r *http.Request, err error) {

	log.Printf("Internal server error: %v, request_id:%v, method:%v, path:%v", err, middleware.GetReqID(r.Context()), r.Method, r.URL.Path)

	writeJSONError(w, http.StatusInternalServerError, "Server encountered an unexpected condition")
}

func (app *application) NotFound(w http.ResponseWriter, r *http.Request) {

	log.Printf("Not found: request_id:%v, method:%v, path:%v", middleware.GetReqID(r.Context()), r.Method, r.URL.Path)

	writeJSONError(w, http.StatusNotFound, "The requested resource could not be found")
}
func (app *application) BadRequest(w http.ResponseWriter, r *http.Request, err error) {

	log.Printf("Bad request: %v, request_id:%v, method:%v, path:%v", err, middleware.GetReqID(r.Context()), r.Method, r.URL.Path)

	writeJSONError(w, http.StatusBadRequest, "The request could not be understood or was missing required parameters")
}
