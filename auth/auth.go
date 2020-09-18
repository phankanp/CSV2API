package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
)

// Creates new api for user
func GenerateAPIKey(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// Hashes password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Compares password
func CheckPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Gets session token from cookie
func GetSessionToken(r *http.Request) (string, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}

	sessionToken := c.Value

	return sessionToken, nil
}

// Gets user email from redis cache
func GetUserEmailFromSessionToken(cache redis.Conn, sessionToken string) (string, error) {
	response, err := redis.String(cache.Do("GET", sessionToken))

	if err != nil {
		return "", err
	}

	return response, nil
}
