package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (app *application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := readBearerToken(r)
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		token, err := app.auth.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			app.unauthorizedError(w, r, errors.New("invalid or expired token"))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			app.unauthorizedError(w, r, errors.New("invalid token claims"))
			return
		}

		subject, ok := claims["sub"]
		if !ok {
			app.unauthorizedError(w, r, errors.New("invalid token subject"))
			return
		}

		userID, err := parseSubjectID(subject)
		if err != nil {
			app.unauthorizedError(w, r, err)
			return
		}

		user, err := app.services.UserSrvc.GetUserByID(r.Context(), userID)
		if err != nil {
			app.unauthorizedError(w, r, errors.New("user not found"))
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func readBearerToken(r *http.Request) (string, error) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return "", errors.New("invalid Authorization header format")
		}

		token := strings.TrimSpace(parts[1])
		if token == "" {
			return "", errors.New("missing bearer token")
		}
		return token, nil
	}

	// Browser WebSocket API custom header yubora olmagani uchun
	// websocket upgrade requestda query token fallback ishlatiladi.
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket") {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token != "" {
			return token, nil
		}
	}

	return "", errors.New("missing Authorization header")
}

func parseSubjectID(subject any) (int64, error) {
	switch value := subject.(type) {
	case float64:
		id := int64(value)
		if id <= 0 {
			return 0, errors.New("invalid token subject")
		}
		return id, nil
	case string:
		id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil || id <= 0 {
			return 0, errors.New("invalid token subject")
		}
		return id, nil
	default:
		return 0, errors.New("invalid token subject")
	}
}
