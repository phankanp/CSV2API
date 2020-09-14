package response

import (
	"net/http"
)

type Error struct {
	Code    int
	Message string
}

var (
	ErrInternal = &Error{
		Code:    http.StatusInternalServerError,
		Message: "Internal server response",
	}
)
