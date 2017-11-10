package ocenet

import (
	"encoding/json"
	"net/http"
)

// JSONApiError sets the status code. The error description and error string
// are written RESTError struct and encoded to JSON, written to response writer.
func JSONApiError(w http.ResponseWriter, statusCode int, errorDesc string, err error) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	resterr := ApiHttpError{
		Error: err.Error(),
		ErrorDescription: errorDesc,
	}
	json.NewEncoder(w).Encode(resterr)
}
