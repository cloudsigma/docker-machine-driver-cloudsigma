package api

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors_Error_messageFormat(t *testing.T) {
	errorResponse := &ErrorResponse{
		Response: &http.Response{
			Request: &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Scheme: "https", Path: "cloudsigma.com/api"},
			},
			StatusCode: 200,
		},
		ErrorElements: []ErrorElement{
			{Message: "first", Type: "permission"},
			{Message: "second"},
		},
	}
	expectedMessage := "GET https://cloudsigma.com/api: 200 [{Message:first Point: Type:permission} {Message:second Point: Type:}]"

	assert.Error(t, errorResponse)
	assert.Equal(t, expectedMessage, errorResponse.Error())
}
