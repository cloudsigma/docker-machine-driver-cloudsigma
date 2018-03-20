package api

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
	client *Client
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewBasicAuthClient("user", "password")
	urlStr, _ := url.Parse(server.URL)
	client.BaseURL = urlStr
}

func teardown() {
	server.Close()
}

func testMethod(t *testing.T, r *http.Request, expected string) {
	if expected != r.Method {
		t.Errorf("Request method = %v, expected %v", r.Method, expected)
	}
}

func TestClient_SetLocationForBaseURL_emptyLocation(t *testing.T) {
	client = NewBasicAuthClient("user", "password")

	client.SetLocationForBaseURL("")

	assert.Equal(t, "https://zrh.cloudsigma.com/api/2.0/", client.BaseURL.String())
}

func TestClient_SetLocationForBaseURL_customLocation(t *testing.T) {
	client = NewBasicAuthClient("user", "password")

	client.SetLocationForBaseURL("wdc")

	assert.Equal(t, "https://wdc.cloudsigma.com/api/2.0/", client.BaseURL.String())
}

func TestClient_CheckResponse_errorElements(t *testing.T) {
	resp := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(strings.NewReader(`[{"error_message":"error"}]`)),
	}
	expected := []ErrorElement{
		{Message: "error"},
	}

	err := CheckResponse(resp).(*ErrorResponse)

	assert.Error(t, err)
	assert.Equal(t, 400, err.Response.StatusCode)
	assert.Equal(t, expected, err.ErrorElements)
}

func TestClient_CheckResponse_noBody(t *testing.T) {
	resp := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}

	err := CheckResponse(resp).(*ErrorResponse)

	assert.Error(t, err)
	assert.Equal(t, 400, err.Response.StatusCode)
	assert.Nil(t, err.ErrorElements)
}
