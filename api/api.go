package api

import (
	"net/http"
)

// A Client manages communication with the CloudSigma API.
type Client struct {
	client *http.Client
}
