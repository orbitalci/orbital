package ocenet

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net/http/httptest"
	"testing"
)
func TestJSONApiError(t *testing.T){
	w := httptest.NewRecorder()
	JSONApiError(w, 400, "missing!", errors.New("test error"))
	expectedRestErr := ApiHttpError{
		Error: "test error",
		ErrorDescription: "missing!",
	}
	res := w.Result()
	if res.StatusCode != 400 {
		t.Errorf("Exected status code %d, got %d", 400, res.StatusCode)
	}
	actionRestErr := ApiHttpError{}
	decoder := json.NewDecoder(res.Body)
	_ = decoder.Decode(&actionRestErr)
	if actionRestErr.Error != expectedRestErr.Error {
		t.Errorf("Actual Error different than Expected. \nActual: %v\nExpected: %v", actionRestErr.Error, expectedRestErr.Error)
	}
	if actionRestErr.ErrorDescription != expectedRestErr.ErrorDescription {
		t.Errorf("Actual Error Description different than Expected. \nActual: %v\nExpected: %v", actionRestErr.ErrorDescription, expectedRestErr.ErrorDescription)
	}
}