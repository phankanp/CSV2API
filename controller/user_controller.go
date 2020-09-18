package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pborman/uuid"
	"github.com/phankanp/csv-to-json/model"
	"github.com/phankanp/csv-to-json/response"
)

// Creates a new user
func (server *Server) Register(w http.ResponseWriter, r *http.Request) {
	user := model.User{}

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	user.Prepare()
	err = user.ValidateInput(server.DB)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	registeredUserAuthKey, err := user.CreateUser(server.DB)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
		return
	}

	response.JsonResponse(w, http.StatusOK, registeredUserAuthKey)
}

// Login user based on email/password
func (server *Server) Login(w http.ResponseWriter, r *http.Request) {
	user := model.User{}

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	email, err := user.CheckCredentials(server.DB, user.Email, user.Password)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusUnauthorized)
		return
	}

	sessionToken := uuid.NewRandom().String()

	_, err = server.Cache.Do("SETEX", sessionToken, "120", email)

	if err != nil {
		response.ErrorResponse(w, err, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: time.Now().Add(120 * time.Second),
	})
}
