package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Basic authorization header for username: "user" and password: "password"
const authorizationHeader = "Basic dXNlcjpwYXNzd29yZA=="

func setup() (client *Client, mux *http.ServeMux, serverURL string, teardown func()) {
	mux = http.NewServeMux()

	apiHandler := http.NewServeMux()
	apiHandler.Handle("/api/2.0/", http.StripPrefix("/api/2.0", mux))
	apiHandler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintln(os.Stderr, "FAIL: Client.BaseURL path prefix is not preserved in the request URL:")
		_, _ = fmt.Fprintln(os.Stderr)
		_, _ = fmt.Fprintln(os.Stderr, "\t"+req.URL.String())
		_, _ = fmt.Fprintln(os.Stderr)
		_, _ = fmt.Fprintln(os.Stderr, "\tDid you accidentally use an absolute endpoint URL rather then relative?")
		http.Error(w, "client.BaseURL path prefix is not preserved in the request URL.", http.StatusInternalServerError)
	})
	server := httptest.NewServer(apiHandler)
	client = NewBasicAuthClient("user", "password")
	client.BaseURL, _ = url.Parse(server.URL + "/api/2.0/")

	return client, mux, server.URL, server.Close
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

func TestClient_CheckResponse_errorWhenUnmarshall(t *testing.T) {
	resp := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(strings.NewReader(`{"error_message":"response is always an array of errors"}`)),
	}

	err := CheckResponse(resp).(*json.UnmarshalTypeError)

	assert.Error(t, err)
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

func TestClient_CheckResponse_noErrorStatusCode(t *testing.T) {
	resp := &http.Response{
		Request:    &http.Request{},
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}

	err := CheckResponse(resp)

	assert.NoError(t, err)
}

func TestClient_Do(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	type foo struct {
		A string
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		_, _ = fmt.Fprint(w, `{"A":"a"}`)
	})
	req, _ := client.NewRequest("GET", ".", nil)
	body := new(foo)

	_, _ = client.Do(req, body)
	expected := &foo{"a"}

	assert.Equal(t, body, expected)
}

func TestClient_Do_httpError(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	})
	req, _ := client.NewRequest("GET", ".", nil)

	resp, err := client.Do(req, nil)

	assert.Error(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestClient_NewBasicAuthClient(t *testing.T) {
	client := NewBasicAuthClient("user", "password")

	assert.Equal(t, "https://zrh.cloudsigma.com/api/2.0/", client.BaseURL.String())
	assert.Equal(t, "docker-machine-driver-cloudsigma", client.UserAgent)
}

func TestClient_NewRequest(t *testing.T) {
	client := NewBasicAuthClient("user", "password")

	req, err := client.NewRequest("GET", "ips/uuid", nil)

	assert.NoError(t, err)
	assert.Equal(t, "https://zrh.cloudsigma.com/api/2.0/ips/uuid", req.URL.String())
}

func TestClient_NewRequest_baseURLWithoutTrailingSlash(t *testing.T) {
	client := NewBasicAuthClient("user", "password")
	client.BaseURL, _ = url.Parse("https://zrh.cloudsigma.com/api/2.0")

	_, err := client.NewRequest("GET", "ips/uuid", nil)

	assert.Error(t, err)
}

func TestClient_NewRequest_invalidRequestURL(t *testing.T) {
	client := NewBasicAuthClient("user", "password")
	client.BaseURL, _ = url.Parse("/")

	_, err := client.NewRequest("GET", ":%31", nil)

	assert.Error(t, err)
}

func TestClient_SetLocationForBaseURL_customLocation(t *testing.T) {
	client := NewBasicAuthClient("user", "password")

	client.SetLocationForBaseURL("wdc")

	assert.Equal(t, "https://wdc.cloudsigma.com/api/2.0/", client.BaseURL.String())
}

func TestClient_SetLocationForBaseURL_emptyLocation(t *testing.T) {
	client := NewBasicAuthClient("user", "password")

	client.SetLocationForBaseURL("")

	assert.Equal(t, "https://zrh.cloudsigma.com/api/2.0/", client.BaseURL.String())
}
