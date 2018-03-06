package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
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

func format(infoMessage string, expected, got interface{}) string {
	return fmt.Sprintf("Info: %v\nExpected: %v\n     Got: %v", infoMessage, expected, got)
}
