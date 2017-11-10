package ocenet

import (
	"encoding/json"
	"net/http"
)

// JSONApiError sets the status code to 400. The error description and error string
// are written RESTError struct and encoded to JSON, written to response writer.
func JSONApiError(w http.ResponseWriter, errorDesc string, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/json")
	resterr := ApiHttpError{
		Error: err,
		ErrorDescription: errorDesc,
	}
	json.NewEncoder(w).Encode(resterr)
}
