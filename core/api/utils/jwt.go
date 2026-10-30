package utils

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	// Refuse the placeholder / weak secrets outside of debug mode.
	if os.Getenv("GIN_MODE") == "release" && (len(secret) < 32 || secret == "your_jwt_secret" || strings.Contains(secret, "change-me")) {
		log.Fatal("JWT_SECRET must be a random value of at least 32 characters when GIN_MODE=release")
	}
	jwtSecret = []byte(secret)
}

// GenerateToken creates a new JWT for the given userID valid for 24 hours
func GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 365 * 10).Unix(), // 10 years
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// VerifyToken parses the token and returns the user_id if valid
func VerifyToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", errors.New("user_id not found in token")
		}
		return userID, nil
	}

	return "", errors.New("invalid token")
}
