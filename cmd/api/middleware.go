package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (app *application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedError(w, r, errors.New("missing Authorization header"))
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			app.unauthorizedError(w, r, errors.New("invalid Authorization header format"))
			return
		}
		tokenString := parts[1]

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
		sub, ok := claims["sub"].(float64)
		if !ok {
			app.unauthorizedError(w, r, errors.New("invalid token subject"))
			return
		}
		userID := int64(sub)

		user, err := app.services.UserSrvc.GetUserByID(r.Context(), userID)
		if err != nil {
			app.unauthorizedError(w, r, errors.New("user not found"))
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
