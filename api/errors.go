package api

import (
	"fmt"
	"net/http"
)

// An ErrorResponse reports one or more errors caused by an API request.
//
// CloudSigma API docs: http://cloudsigma-docs.readthedocs.io/en/2.14/errors.html
type ErrorResponse struct {
	Response      *http.Response // HTTP response that caused this error.
	ErrorElements []ErrorElement
}

type ErrorElement struct {
	Message string `json:"error_message"`
	Point   string `json:"error_point,omitempty"`
	Type    string `json:"error_type"`
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %+v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.ErrorElements)
}
