package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func ErrorResponse(w http.ResponseWriter, err error, message string, code int) {
	errObj := HTTPError{
		Error:   err.Error(),
		Message: message,
		Code:    code,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	err = json.NewEncoder(w).Encode(errObj)

	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}
