// internal/auth/jwt.go
package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Sub   string `json:"sub"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	return []byte(secret)
}

func GenerateJWT(userID string, isAdmin bool) (string, error) {
	claims := &Claims{
		Sub:   userID,
		Admin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt:	jwt.NewNumericDate()
		},
	}
}
