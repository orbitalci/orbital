package ocenet

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/shankj3/ocelot/util"
	"net/http/httptest"
	"testing"
)
func TestJSONApiError(t *testing.T){
	w := httptest.NewRecorder()
	JSONApiError(w, 400, "missing!", errors.New("test error"))
	expectedRestErr := ApiHttpError{
		Error: "test error",
		ErrorDescription: "missing!",
		Status: 400,
	}
	res := w.Result()
	if res.StatusCode != 400 {
		t.Error(util.IntFormatErrors("status code", 400, res.StatusCode))
	}
	actionRestErr := ApiHttpError{}
	decoder := json.NewDecoder(res.Body)
	_ = decoder.Decode(&actionRestErr)
	if actionRestErr.Error != expectedRestErr.Error {
		t.Error(util.StrFormatErrors("error", expectedRestErr.Error, actionRestErr.Error))
	}
	if actionRestErr.ErrorDescription != expectedRestErr.ErrorDescription {
		t.Error(util.StrFormatErrors("error description", expectedRestErr.ErrorDescription, actionRestErr.ErrorDescription))
	}
}