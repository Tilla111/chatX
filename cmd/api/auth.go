package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func getUserIDFromRequest(r *http.Request) (int64, error) {
	raw := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if raw == "" {
		raw = strings.TrimSpace(r.URL.Query().Get("user_id"))
	}
	if raw == "" {
		return 0, errors.New("X-User-ID header is required")
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("X-User-ID must be a positive integer")
	}

	return id, nil
}

func (app *application) requireUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		app.unauthorizedError(w, r, err)
		return 0, false
	}

	return userID, true
}
