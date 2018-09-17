package slack

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// MakeFakePoster will create a FakePoster instance for you with the desired resopnse code and response body set
func MakeFakePoster(desiredStatusCode int, desiredReturnBody string) *FakePoster {
	return &FakePoster{ResponseCode: desiredStatusCode, ResponseBody: desiredReturnBody}
}

// FakePoster is for testing anything related to posting slack webhooks
//   When you use it instead of http.DefaultClient in ThrowStatusWebhook the body posted to the client
//   will be saved to the PostBody attribute. FakePoster will reutrn the ResponseCode and ResponseBody
//   it will also save a list of urls that have been posted to it
type FakePoster struct {
	// the data that was posted to FakePoster will be posted here
	PostBody []byte
	// response code to return
	ResponseCode int
	// body to return in response
	ResponseBody string
	// urls that have been posted
	PostedUrls []string
}

func (f *FakePoster) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	f.PostedUrls = append(f.PostedUrls, url)
	resp = &http.Response{}
	f.PostBody, err = ioutil.ReadAll(body)
	if err != nil {
		return
	}
	resp.StatusCode = f.ResponseCode
	resp.Body = ioutil.NopCloser(strings.NewReader(f.ResponseBody))
	return
}
