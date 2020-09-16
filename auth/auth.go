package auth

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func GenerateAPIKey(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetSessionToken(r *http.Request) (string, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}

	sessionToken := c.Value

	return sessionToken, nil
}

func GetUserEmailFromSessionToken(cache redis.Conn, sessionToken string) (string, error) {
	response, err := redis.String(cache.Do("GET", sessionToken))

	if err != nil {
		return "", err
	}

	return response, nil
}
