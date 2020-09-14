package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func JsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)

	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}
