package ocenet

import (
	"encoding/json"
	"github.com/shankj3/ocelot/util/ocelog"
	"net/http"
)

// JSONApiError sets the status code. The error description and error string
// are written RESTError struct and encoded to JSON, written to response writer.
// Also logs the error using ocelog
func JSONApiError(w http.ResponseWriter, statusCode int, errorDesc string, err error) {
	ocelog.IncludeErrField(err).Error(errorDesc)
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	restErr := ApiHttpError{
		Error: err.Error(),
		ErrorDescription: errorDesc,
		Status: statusCode,
	}
	json.NewEncoder(w).Encode(restErr)
}
