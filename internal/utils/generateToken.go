package utils

import (
	. "awesomeProject15/db-service-homework-Ppolyak/internal/database"
	"github.com/dgrijalva/jwt-go"
	"time"
)

func GenerateToken(user User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("your_secret_key"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
