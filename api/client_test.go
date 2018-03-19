package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func format(infoMessage string, expected, got interface{}) string {
	return fmt.Sprintf("Info: %v\nExpected: %v\n     Got: %v", infoMessage, expected, got)
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
