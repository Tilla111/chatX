package auth

import "github.com/golang-jwt/jwt/v5"

type JWTAuthenticator struct {
	secretKey string
	aud       string
	iss       string
}

func NewJWTAuthenticator(secretKey, aud, iss string) *JWTAuthenticator {
	return &JWTAuthenticator{
		secretKey: secretKey,
		aud:       aud,
		iss:       iss,
	}
}

func (j *JWTAuthenticator) CreateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTAuthenticator) ValidateToken(tokenString string) (*jwt.Token, error) {
	options := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	}

	if j.aud != "" {
		options = append(options, jwt.WithAudience(j.aud))
	}
	if j.iss != "" {
		options = append(options, jwt.WithIssuer(j.iss))
	}

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	}, options...)
}
