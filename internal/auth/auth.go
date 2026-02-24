package auth

import "github.com/golang-jwt/jwt/v5"

type AuthService interface {
	CreateToken(Claims jwt.Claims) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
}
