package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultLocation = "zrh"
	defaultBaseURL  = "https://" + defaultLocation + ".cloudsigma.com/api/2.0/"
	userAgent       = "docker-machine-driver-cloudsigma"

	mediaType = "application/json"
)

// A Client manages communication with the CloudSigma API.
type Client struct {
	client *http.Client // HTTP client used to communicate with the API.

	BaseURL   *url.URL // Base URL for API requests. BaseURL should always be specified with a trailing slash.
	UserAgent string   // User agent used when communicating with the CloudSigma API.

	Username string // Username for CloudSigma API (user email).
	Password string // Password for CloudSigma API.

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	LibraryDrives *LibraryDrivesService
}

type service struct {
	client *Client
}

// NewBasicAuthClient returns a new CloudSigma API client. To use API methods provide username (your email)
// and password.
func NewBasicAuthClient(username, password string) *Client {
	httpClient := http.DefaultClient
	baseUrl, _ := url.Parse(defaultBaseURL)

	c := &Client{client: httpClient, BaseURL: baseUrl, UserAgent: userAgent, Username: username, Password: password}
	c.common.client = c
	c.LibraryDrives = (*LibraryDrivesService)(&c.common)
	return c
}

// NewRequest creates an API request. A relative URL can be provided in urlStr, in which case it is resolved
// relative to the BaseURL of the Client. Relative URLs should always be specified without a preceding slash.
// If specified, the value pointed to by body is JSON encoded and included as the request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if body != nil {
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Content-Type", mediaType)
	req.Header.Set("Accept", mediaType)
	req.Header.Set("User-Agent", c.UserAgent)
	return req, nil
}

// Do sends an API request and returns the API response. The API response is JSON decoded and stored in
// the value pointed to by v, or returned as an error if an API error has occurred.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}()

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				return nil, err
			}
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err != nil {
				return nil, err
			}
		}
	}

	return resp, err
}

// CheckResponse checks the API response for errors, and returns them if present. A response is considered
// an error if it has a status code outside the 200 range.
func CheckResponse(resp *http.Response) error {
	if code := resp.StatusCode; code >= 200 && code <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: resp}
	data, err := ioutil.ReadAll(resp.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, &errorResponse.ErrorElements)
		if err != nil {
			return err
		}
	}
	return errorResponse
}
